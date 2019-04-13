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
	pb "github.com/HayoVanLoon/protoworkflow-genproto/bobsknobshop/messaging/v1"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
)

const (
	port              = "8080"
)

type server struct {
}

func (server) PostMessage(context.Context, *pb.PostMessageRequest) (*pb.PostMessageResponse, error) {
	panic("implement me")
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
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterMessagingServer(s, &server{})

	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
