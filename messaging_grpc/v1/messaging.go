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
	messageLimit        = 10
)

type server struct {
	services map[string]string
}

func newServer(services map[string]string) *server {
	return &server{services}
}

// Provides a storage client
func (s server) getStorageClient() (storagepb.StorageClient, func(), error) {
	conn, err := s.getConn(storageService)
	if err != nil {
		log.Print("ERROR: could not open connection to storage")
		return nil, nil, err
	}
	return storagepb.NewStorageClient(conn), closeConnFn(conn), err
}

// Provides a connection-closing function
func closeConnFn(conn *grpc.ClientConn) func() {
	return func() {
		if err := conn.Close(); err != nil {
			log.Printf("WARN: error closing connection: %v", err)
		}
	}
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

func (s server) storeMessage(m *pb.CustomerMessage) (string, error) {
	key := createKey(m)

	data, _ := proto.Marshal(m)
	r := &storagepb.PostObjectRequest{Key: key, Data: data}

	c, closeConn, err := s.getStorageClient()
	if err != nil {
		return "", err
	}
	defer closeConn()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.PostObject(ctx, r)
	if err != nil {
		log.Printf("WARN: storing message: %v", err)
		return "", err
	} else if resp.GetName() != "" {
		log.Printf("DEBUG: stored message %v", m.Body)
	} else {
		log.Printf("WARN: message already stored %v", m.Body)
	}

	return resp.GetName(), err
}

func (s server) getMessages(cat pb.MessageCategory, st pb.Status, l int32) (msgs []*pb.CustomerMessage, err error) {
	query := &storagepb.Key{
		IndexedValues: []*storagepb.Key_Part{
			{Key: "category", Value: cat.String()},
			{Key: "status", Value: st.String()},
		},
	}

	r := &storagepb.GetObjectRequest{Keys: []*storagepb.Key{query}, Limit: l}

	c, closeConn, err := s.getStorageClient()
	if err != nil {
		return nil, err
	}
	defer closeConn()

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

func (s server) mutateMessage(oldM, newM *pb.CustomerMessage) (bool, error) {
	oldKey := createKey(oldM)
	newKey := createKey(newM)

	oldData, _ := proto.Marshal(oldM)
	newData, _ := proto.Marshal(newM)
	r := &storagepb.MutateObjectRequest{OldKey: oldKey, NewKey: newKey, OldData: oldData, NewData: newData}

	c, closeConn, err := s.getStorageClient()
	if err != nil {
		return false, err
	}
	defer closeConn()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.MutateObject(ctx, r)
	if err != nil {
		log.Printf("WARN: could not mutated message \"%s\"", oldM.Body)
		return false, err
	} else if !resp.Success {
		log.Printf("WARN: could not mutated message \"%s\"", oldM.Body)
	} else {
		log.Printf("DEBUG: mutated message \"%s\"", oldM.Body)
	}

	return resp.Success, err
}

func (s server) PostMessage(ctx context.Context, r *pb.PostMessageRequest) (resp *pb.PostMessageResponse, err error) {
	m := r.GetCustomerMessage()
	m.Status = pb.Status_TO_DO

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
	var name string
	for i := -1; (err != nil || i < 0) && i < maxRetries; i += 1 {
		name, err = s.storeMessage(m)
	}
	if err != nil {
		log.Printf("ERROR: error while storing message: %s", err)
		return
	} else if name == "" {
		log.Printf("WARN: could not store message: %s", m.Body)
		return &pb.PostMessageResponse{}, nil
	}

	m.Name = name
	return &pb.PostMessageResponse{Message: &pb.PostMessageResponse_CustomerMessage{m}}, nil
}

// Claims a message.
func (s server) claimFirst(msgs []*pb.CustomerMessage) (m *pb.CustomerMessage, err error) {
	for i := 0; m == nil && i < len(msgs); i += 1 {
		oldM := *msgs[i]
		m = msgs[i]
		m.Status = pb.Status_IN_PROCESS
		ok, err := s.mutateMessage(&oldM, m)
		if !ok || err != nil {
			m = nil
		}
	}
	return
}

func (s server) GetQuestion(context.Context, *pb.GetQuestionRequest) (resp *pb.GetQuestionResponse, err error) {
	for {
		msgs, err := s.getMessages(pb.MessageCategory_QUESTION, pb.Status_TO_DO, messageLimit)
		if err != nil {
			log.Printf("WARN: error retrieving messages: %s", err)
			return nil, err
		}
		if len(msgs) == 0 {
			break
		}

		m, err := s.claimFirst(msgs)
		if err != nil {
			log.Printf("WARN: error getting question: %s", err)
			return nil, err
		} else {
			return &pb.GetQuestionResponse{
				Message: &pb.GetQuestionResponse_CustomerMessage{m},
			}, err
		}
	}

	return &pb.GetQuestionResponse{}, nil
}

func (s server) GetComplaint(context.Context, *pb.GetComplaintRequest) (*pb.GetComplaintResponse, error) {
	for {
		msgs, err := s.getMessages(pb.MessageCategory_COMPLAINT, pb.Status_TO_DO, messageLimit)
		if err != nil {
			log.Printf("WARN: error retrieving messages: %s", err)
			return nil, err
		}
		if len(msgs) == 0 {
			break
		}

		m, err := s.claimFirst(msgs)
		if err != nil {
			log.Printf("WARN: error getting question: %s", err)
			return nil, err
		} else {
			return &pb.GetComplaintResponse{
				Message: &pb.GetComplaintResponse_CustomerMessage{m},
			}, err
		}
	}

	return &pb.GetComplaintResponse{}, nil
}

func (s server) GetFeedback(context.Context, *pb.GetFeedbackRequest) (resp *pb.GetFeedbackResponse, err error) {
	for {
		msgs, err := s.getMessages(pb.MessageCategory_FEEDBACK, pb.Status_TO_DO, messageLimit)
		if err != nil {
			log.Printf("WARN: error retrieving messages: %s", err)
			return nil, err
		}
		if len(msgs) == 0 {
			break
		}

		m, err := s.claimFirst(msgs)
		if err != nil {
			log.Printf("WARN: error getting question: %s", err)
			return nil, err
		} else {
			return &pb.GetFeedbackResponse{
				Message: &pb.GetFeedbackResponse_CustomerMessage{m},
			}, err
		}
	}

	return &pb.GetFeedbackResponse{}, nil
}

func (server) MoveMessage(context.Context, *pb.MoveMessageRequest) (*pb.MoveMessageResponse, error) {
	panic("implement me")
}

func (server) UpdateStatus(context.Context, *pb.UpdateStatusRequest) (*pb.UpdateStatusResponse, error) {
	panic("implement me")
}

func (s server) GetMessage(_ context.Context, req *pb.GetMessageRequest) (*pb.GetMessageResponse, error) {
	c, closeConn, err := s.getStorageClient()
	if err != nil {
		return nil, err
	}
	defer closeConn()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	key := req.GetName()
	r := &storagepb.GetObjectRequest{Keys: []*storagepb.Key{{Name: key}}, Limit: 1}

	objResp, err := c.GetObject(ctx, r)
	if err != nil {
		log.Printf("WARN: error fetching message '%s'", key)
		return nil, err
	}

	m := &pb.CustomerMessage{}
	err = proto.Unmarshal(objResp.Data[0], m)
	if err != nil {
		log.Printf("WARN: error unmarshalling message '%s'", key)
		return nil, err
	}

	return &pb.GetMessageResponse{Message: &pb.GetMessageResponse_CustomerMessage{m}}, nil
}

func (server) SearchMessages(context.Context, *pb.SearchMessagesRequest) (*pb.SearchMessagesResponse, error) {
	panic("implement me")
}

func (s server) DeleteMessage(ctx context.Context, req *pb.DeleteMessageRequest) (*empty.Empty, error) {
	r := &storagepb.DeleteObjectRequest{Keys: []*storagepb.Key{{Name: req.Name}}}

	c, closeConn, err := s.getStorageClient()
	if err != nil {
		return nil, err
	}
	defer closeConn()

	_, err = c.DeleteObject(ctx, r)
	if err != nil {
		log.Printf("WARN: error getting stored message: %v", err)
		return nil, err
	}

	return &empty.Empty{}, err
}

func main() {
	var port = flag.String("port", defaultPort, "port to listen on")
	var storageHost = flag.String("storage-host", storageService, "storage service")
	var storagePort = flag.String("storage-port", defaultPort, "storage service port")
	var categorisingHost = flag.String("categorising-host", categorisingService, "categorising service")
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
