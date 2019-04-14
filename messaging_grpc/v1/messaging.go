/*
 * Copyright 2019 Hayo van Loon
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package main

import (
	"flag"
	"fmt"
	categorisingpb "github.com/HayoVanLoon/protoworkflow-genproto/bobsknobshop/categorising/v1"
	pb "github.com/HayoVanLoon/protoworkflow-genproto/bobsknobshop/messaging/v1"
	storagepb "github.com/HayoVanLoon/protoworkflow-genproto/bobsknobshop/storage/v1"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"strconv"
	"time"
)

const (
	storageService      = "storage-service"
	categorisingService = "categorising-service"
	defaultPort         = "8080"
	maxRetries          = 3
)

type server struct {
	services map[string]string
}

func newServer(services map[string]string) *server {
	return &server{services}
}

// Establishes a connection to a service
func (s server) getConn(service string) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(s.services[service], grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("did not connect: %v", err)
	}
	return conn, nil
}

// Creates a key for the message
func createKey(m *pb.CustomerMessage) *storagepb.Key {
	return &storagepb.Key{
		Parts: []*storagepb.Key_Part{
			{Key: "timestamp", Value: strconv.FormatInt(m.Timestamp, 10)},
			{Key: "id", Value: m.Sender.Name},
		},
		IndexedValues: []*storagepb.Key_Part{
			{Key: "category", Value: m.GetCategory().String()},
			{Key: "status", Value: m.GetStatus().String()},
		},
	}
}

func (s server) getCategory(m *pb.CustomerMessage) (cat pb.MessageCategory, err error) {
	r := &categorisingpb.GetCategoryRequest{Text: m.Body}

	conn, err := s.getConn(categorisingService)
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("WARN: error closing connection: %v", err)
		}
	}()

	c := categorisingpb.NewCategorisingClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.GetCategory(ctx, r)
	if err == nil {
		cat = resp.GetCategory()
	}

	return
}

func (s server) storeMessage(m *pb.CustomerMessage) error {
	key := createKey(m)

	data, _ := proto.Marshal(m)
	r := &storagepb.PostObjectRequest{Key: key, Data: data}

	conn, err := s.getConn(storageService)
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("WARN: error closing connection: %v", err)
		}
	}()

	c := storagepb.NewStorageClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.PostObject(ctx, r)
	if err != nil {
		log.Printf("WARN: storing message: %v", err)
	} else if resp.Success {
		log.Printf("DEBUG: stored message %v", m.Body)
	} else {
		log.Printf("WARN: message already stored %v", m.Body)
	}

	return err
}

func (s server) getMessages(cat pb.MessageCategory, st pb.Status) (msgs []*pb.CustomerMessage, err error) {
	query := &storagepb.Key{
		IndexedValues: []*storagepb.Key_Part{
			{Key: "category", Value: cat.String()},
			{Key: "status", Value: st.String()},
		},
	}

	r := &storagepb.GetObjectRequest{Keys: []*storagepb.Key{query}}

	conn, err := s.getConn(storageService)
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("WARN: error closing connection: %v", err)
		}
	}()

	c := storagepb.NewStorageClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.GetObject(ctx, r)
	if err != nil {
		log.Printf("WARN: error getting stored message: %v", err)
		return
	}

	for i := 0; i < len(resp.Data); i += 1 {
		m := &pb.CustomerMessage{}
		_ = proto.Unmarshal(resp.Data[i], m)
		msgs = append(msgs, m)
	}

	return
}

func (s server) mutateMessage(old *pb.CustomerMessage, new *pb.CustomerMessage) error {
	key := createKey(old)

	oldData, _ := proto.Marshal(old)
	newData, _ := proto.Marshal(new)
	r := &storagepb.MutateObjectRequest{Key: key, Old: oldData, New: newData}

	conn, err := s.getConn(storageService)
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("WARN: error closing connection: %v", err)
		}
	}()

	c := storagepb.NewStorageClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = c.MutateObject(ctx, r)
	log.Printf("DEBUG: mutated message %v-%v-%v", old.Timestamp, old.Sender.Name, old.Body)

	return err
}

func (s server) PostMessage(ctx context.Context, r *pb.PostMessageRequest) (resp *pb.PostMessageResponse, err error) {
	m := r.GetCustomerMessage()
	m.Status = pb.Status_TO_DO
	m.Timestamp = time.Now().UnixNano() / 1000000

	// set category
	var cat pb.MessageCategory
	for i := -1; (err != nil || i < 0) && i < maxRetries; i += 1 {
		cat, err = s.getCategory(m)
	}
	if err != nil {
		log.Printf("WARN: could not get category: %s", err)
		return
	}
	m.Category = cat

	// store message
	for i := -1; (err != nil || i < 0) && i < maxRetries; i += 1 {
		err = s.storeMessage(m)
	}
	if err != nil {
		log.Printf("ERROR: could not store message: %s", err)
		return
	}

	return
}

// Claims a message.
func (s server) claimFirst(msgs []*pb.CustomerMessage) (m *pb.CustomerMessage, err error) {
	for i := 0; m == nil && i < len(msgs); i += 1 {
		m = msgs[i]
		m.Status = pb.Status_IN_PROCESS
		err = s.storeMessage(m)
		if err != nil {
			m = nil
		}
	}
	return
}

func (s server) GetQuestion(context.Context, *pb.GetQuestionRequest) (resp *pb.GetQuestionResponse, err error) {
	msgs, err := s.getMessages(pb.MessageCategory_QUESTION, pb.Status_TO_DO)
	if err != nil {
		log.Printf("WARN: error retrieving messages: %s", err)
	}

	m, err := s.claimFirst(msgs)
	if err == nil {
		resp = &pb.GetQuestionResponse{Message: &pb.GetQuestionResponse_CustomerMessage{m}}
	}

	return
}

func (s server) GetComplaint(context.Context, *pb.GetComplaintRequest) (resp *pb.GetComplaintResponse, err error) {
	msgs, err := s.getMessages(pb.MessageCategory_QUESTION, pb.Status_TO_DO)
	if err != nil {
		log.Printf("WARN: error retrieving messages: %s", err)
	}

	m, err := s.claimFirst(msgs)
	if m != nil {
		resp = &pb.GetComplaintResponse{Message: &pb.GetComplaintResponse_CustomerMessage{m}}
	}

	return
}

func (s server) GetFeedback(context.Context, *pb.GetFeedbackRequest) (resp *pb.GetFeedbackResponse, err error) {
	msgs, err := s.getMessages(pb.MessageCategory_FEEDBACK, pb.Status_TO_DO)
	if err != nil {
		log.Printf("WARN: error retrieving messages: %s", err)
	}

	m, err := s.claimFirst(msgs)
	if m != nil {
		resp = &pb.GetFeedbackResponse{Message: &pb.GetFeedbackResponse_CustomerMessage{m}}
	}

	return
}

func (server) MoveMessage(context.Context, *pb.MoveMessageRequest) (*pb.MoveMessageResponse, error) {
	panic("implement me")
}

func (server) UpdateStatus(context.Context, *pb.UpdateStatusRequest) (*pb.UpdateStatusResponse, error) {
	panic("implement me")
}

func (server) GetMessage(context.Context, *pb.GetMessageRequest) (*pb.GetMessageResponse, error) {
	panic("implement me")
}

func (server) SearchMessages(context.Context, *pb.SearchMessagesRequest) (*pb.SearchMessagesResponse, error) {
	panic("implement me")
}

func (server) DeleteMessage(context.Context, *pb.DeleteMessageRequest) (*empty.Empty, error) {
	panic("implement me")
}

func main() {
	var port = flag.String("port", defaultPort, "port to listen on")
	var storageHost = flag.String("storage", storageService, "storage service")
	var storagePort = flag.String("storage-port", defaultPort, "storage service port")
	var categorisingHost = flag.String("categorising-service", categorisingService, "categorising service")
	var categorisingPort = flag.String("categorising-port", defaultPort, "categorising service port")
	flag.Parse()

	lis, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterMessagingServer(s, newServer(map[string]string{
		storageService:      *storageHost + ":" + *storagePort,
		categorisingService: *categorisingHost + ":" + *categorisingPort,
	}))

	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
