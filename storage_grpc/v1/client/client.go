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
	pb "github.com/HayoVanLoon/protoworkflow-genproto/bobsknobshop/storage/v1"
	"google.golang.org/grpc"
	"log"
	"time"
)

const (
	host = "localhost"
	defaultPort = "8080"
)

// Establishes a connection to the service
func getConn(port string) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(fmt.Sprintf("%v:%v", host, port), grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("did not connect: %v", err)
	}
	return conn, nil
}

func createObject(port string, key *pb.Key, m string) error {
	r := &pb.CreateObjectRequest{Key: key, Data: []byte(m)}

	conn, err := getConn(port)
	defer func() {
		if err := conn.Close(); err != nil {
			log.Panicf("error closing connection: %v", err)
		}
	}()

	c := pb.NewStorageClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.CreateObject(ctx, r)

	log.Printf("Create %v\n", resp)

	return err
}

func getObject(port string, key *pb.Key) error {
	r := &pb.GetObjectRequest{Keys: []*pb.Key{key}}

	conn, err := getConn(port)
	defer func() {
		if err := conn.Close(); err != nil {
			log.Panicf("error closing connection: %v", err)
		}
	}()

	c := pb.NewStorageClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.GetObject(ctx, r)

	log.Printf("Get %v\n", resp)

	return err
}

func deleteObject(port string, key *pb.Key) error {
	r := &pb.DeleteObjectRequest{Keys: []*pb.Key{key}}

	conn, err := getConn(port)
	defer func() {
		if err := conn.Close(); err != nil {
			log.Panicf("error closing connection: %v", err)
		}
	}()

	c := pb.NewStorageClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.DeleteObject(ctx, r)

	log.Printf("Delete: %v\n", resp)

	return err
}

// Fires a few requests to the service
func examples(port string) {
	key := &pb.Key{Parts: []*pb.Key_Part{{Key: "foo", Value: "1"}, {Key: "bar", Value: "2"}}}
	key2 := &pb.Key{Parts: []*pb.Key_Part{{Key: "foo", Value: "3"}, {Key: "bar", Value: "4"}}}
	query := &pb.Key{IndexedValues: []*pb.Key_Part{{Key: "foo", Value: "*"}, {Key: "bar", Value: "*"}}}

	key3 := &pb.Key{
		Parts: []*pb.Key_Part{
			{Key: "timestamp", Value: "0"},
			{Key: "id", Value: "test1234"},
		},
		IndexedValues: []*pb.Key_Part{
			{Key: "category", Value: "QUESTION"},
			{Key: "status", Value: "TO_DO"},
		},
	}
	key4 := &pb.Key{
		Parts: []*pb.Key_Part{
			{Key: "timestamp", Value: "1"},
			{Key: "id", Value: "test1234"},
		},
		IndexedValues: []*pb.Key_Part{
			{Key: "category", Value: "QUESTION"},
			{Key: "status", Value: "TO_DO"},
		},
	}
	query2 := &pb.Key{IndexedValues: []*pb.Key_Part{
		{Key: "category", Value: "*"},
		{Key: "status", Value: "*"},
	}}

	// _ = getObject(query2)

	_ = createObject(port, key, "bla")
	_ = createObject(port, key2, "blue")
	_ = createObject(port, key3, "I have lots of questions. What's the meaning of life?")
	_ = createObject(port, key4, "Bananas are yellow.")

	_ = getObject(port, key)
	_ = getObject(port, query)

	_ = deleteObject(port, key)

	_ = getObject(port, key2)
	_ = getObject(port, query2)

	// clean up
	_ = deleteObject(port, key2)
	_ = deleteObject(port, key3)
	_ = getObject(port, query2)
	_ = deleteObject(port, key4)
}

func main() {
	var port = flag.String("port", defaultPort, "port to listen on")
	flag.Parse()

	examples(*port)
	query := &pb.Key{IndexedValues: []*pb.Key_Part{
		{Key: "category", Value: "*"},
		{Key: "status", Value: "*"},
	}}

	_ = getObject(*port, query)
}
