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
	"context"
	"flag"
	"fmt"
	contactpb "github.com/HayoVanLoon/protoworkflow-genproto/bobsknobshop/contact/v1"
	pb "github.com/HayoVanLoon/protoworkflow-genproto/bobsknobshop/messaging/v1"
	"google.golang.org/grpc"
	"log"
	"time"
)

const (
	host = "localhost"
	defaultPort = "8080"
)

func getConn(port string) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(host+":"+port, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("did not connect: %v", err)
	}
	return conn, nil
}

func postMessage(port string, m *pb.CustomerMessage) error {
	r := &pb.PostMessageRequest{Message: &pb.PostMessageRequest_CustomerMessage{m}}

	conn, err := getConn(port)
	defer func() {
		if err := conn.Close(); err != nil {
			log.Panicf("error closing connection: %v", err)
		}
	}()

	c := pb.NewMessagingClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.PostMessage(ctx, r)

	log.Printf("%v\n", resp)

	return err
}

func getQuestion(port string) error {
	r := &pb.GetQuestionRequest{}

	conn, err := getConn(port)
	defer func() {
		if err := conn.Close(); err != nil {
			log.Panicf("error closing connection: %v", err)
		}
	}()

	c := pb.NewMessagingClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.GetQuestion(ctx, r)

	log.Printf("%v\n", resp)
	if err != nil {
		log.Printf("Error: %v\n", err)
	}

	return err
}

func main() {
	var port = flag.String("port", defaultPort, "message service port")
	flag.Parse()

	question := &pb.CustomerMessage{
		Body:   "I have a question about this product",
		Sender: &contactpb.Sender{Name: "test1234"},
	}
	complaint := &pb.CustomerMessage{
		Body:   "The knob is too jolly. This does not please me.",
		Sender: &contactpb.Sender{Name: "test1234"},
	}

	_ = postMessage(*port, question)
	_ = postMessage(*port, complaint)
	_ = getQuestion(*port)
}
