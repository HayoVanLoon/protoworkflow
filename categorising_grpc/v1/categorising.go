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
	"cloud.google.com/go/language/apiv1"
	"flag"
	pb "github.com/HayoVanLoon/protoworkflow-genproto/bobsknobshop/categorising/v1"
	messagepb "github.com/HayoVanLoon/protoworkflow-genproto/bobsknobshop/messaging/v1"
	"golang.org/x/net/context"
	languagepb "google.golang.org/genproto/googleapis/cloud/language/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
)

const (
	defaultPort       = "8080"
	questionThreshold = 0
	feedbackThreshold = .6
)

type server struct {
}

func (s *server) GetCategory(ctx context.Context, r *pb.GetCategoryRequest) (*pb.GetCategoryResponse, error) {
	client, err := language.NewClient(ctx)
	if err != nil {
		log.Printf("error: failed to create nlp client: %v", err)
		return nil, err
	}

	sentiment, err := client.AnalyzeSentiment(ctx, &languagepb.AnalyzeSentimentRequest{
		Document: &languagepb.Document{
			Source: &languagepb.Document_Content{Content: r.Text},
			Type:   languagepb.Document_PLAIN_TEXT,
		},
		EncodingType: languagepb.EncodingType_UTF8,
	})
	if err != nil {
		log.Printf("warning: failed to analyze text: %v", err)
		return nil, err
	}
	score := sentiment.DocumentSentiment.Score

	log.Printf("info: message was scored with %v", score)

	resp := &pb.GetCategoryResponse{}
	if score < questionThreshold {
		resp.Category = messagepb.MessageCategory_COMPLAINT
	} else if score < feedbackThreshold {
		resp.Category = messagepb.MessageCategory_QUESTION
	} else {
		resp.Category = messagepb.MessageCategory_FEEDBACK
	}

	return resp, nil
}

func main() {
	var port = flag.String("port", defaultPort, "port to listen on")
	flag.Parse()

	lis, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterCategorisingServer(s, &server{})

	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
