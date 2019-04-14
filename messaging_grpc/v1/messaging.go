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

func (s server) getCategory(m *pb.CustomerMessage) (cat pb.MessageCategory, err error) {
	r := &categorisingpb.GetCategoryRequest{Text: m.Body}

	conn, err := s.getConn(categorisingService)
	defer func() {
		if err := conn.Close(); err != nil {
			log.Panicf("error closing connection: %v", err)
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

	r := &storagepb.PostObjectRequest{Key: key, Data: []byte(m.Body)}

	conn, err := s.getConn(storageService)
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("warning: error closing connection: %v", err)
		}
	}()

	c := storagepb.NewStorageClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.PostObject(ctx, r)

	log.Printf("%v\n", resp)

	return err
}

func createKey(m *pb.CustomerMessage) *storagepb.Key {
	return &storagepb.Key{
		Parts: []*storagepb.Key_Part{
			{Key: "timestamp", Value: strconv.FormatInt(m.Timestamp, 10)},
			{Key: "id", Value: m.Sender.Name},
			{Key: "category", Value: m.GetCategory().String()},
			{Key: "status", Value: m.GetStatus().String()},
		},
	}
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
		log.Printf("warning: could not get category: %s", err)
		return
	}
	m.Category = cat

	// store message
	for i := -1; (err != nil || i < 0) && i < maxRetries; i += 1 {
		err = s.storeMessage(m)
	}
	if err != nil {
		log.Printf("error: could not store message: %s", err)
		return
	}

	return
}

func (server) GetQuestion(context.Context, *pb.GetQuestionRequest) (*pb.GetQuestionResponse, error) {
	panic("implement me")
}

func (server) GetComplaint(context.Context, *pb.GetComplaintRequest) (*pb.GetComplaintResponse, error) {
	panic("implement me")
}

func (server) GetFeedback(context.Context, *pb.GetFeedbackRequest) (*pb.GetFeedbackResponse, error) {
	panic("implement me")
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
