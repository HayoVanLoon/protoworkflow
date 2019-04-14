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
	"fmt"
	pb "github.com/HayoVanLoon/protoworkflow-genproto/bobsknobshop/storage/v1"
	"google.golang.org/grpc"
	"log"
	"time"
)

const (
	host = "localhost"
	port = 8080
)

// Establishes a connection to the service
func getConn() (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(fmt.Sprintf("%v:%v", host, port), grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("did not connect: %v", err)
	}
	return conn, nil
}

func postObject(key *pb.Key, m string) error {
	r := &pb.PostObjectRequest{Key: key, Data: []byte(m)}

	conn, err := getConn()
	defer func() {
		if err := conn.Close(); err != nil {
			log.Panicf("error closing connection: %v", err)
		}
	}()

	c := pb.NewStorageClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.PostObject(ctx, r)

	log.Printf("%v\n", resp)

	return err
}

func getObject(key *pb.Key) error {
	r := &pb.GetObjectRequest{Keys: []*pb.Key{key}}

	conn, err := getConn()
	defer func() {
		if err := conn.Close(); err != nil {
			log.Panicf("error closing connection: %v", err)
		}
	}()

	c := pb.NewStorageClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.GetObject(ctx, r)

	log.Printf("%v\n", resp)

	return err
}

func deleteObject(key *pb.Key) error {
	r := &pb.DeleteObjectRequest{Keys: []*pb.Key{key}}

	conn, err := getConn()
	defer func() {
		if err := conn.Close(); err != nil {
			log.Panicf("error closing connection: %v", err)
		}
	}()

	c := pb.NewStorageClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.DeleteObject(ctx, r)

	log.Printf("%v\n", resp)

	return err
}

// Fires a few requests to the service
func examples() {
	key := &pb.Key{Parts:[]*pb.Key_Part{{Key: "foo", Value:"1"}, {Key: "bar", Value:"2"}}}
	key2 := &pb.Key{Parts:[]*pb.Key_Part{{Key: "foo", Value:"3"}, {Key: "bar", Value:"4"}}}
	query := &pb.Key{Parts: []*pb.Key_Part{{Key: "foo", Value:"*"}, {Key: "bar", Value:"*"}}}

	key3 := &pb.Key{Parts:[]*pb.Key_Part{
		{Key: "timestamp", Value: "0"},
		{Key: "id", Value: "test1234"},
		{Key: "category", Value: "QUESTION"},
		{Key: "status", Value: "TO_DO"},
	}}
	query2 := &pb.Key{Parts:[]*pb.Key_Part{
		{Key: "timestamp", Value: "*"},
		{Key: "id", Value:"*"},
		{Key: "category", Value:"*"},
		{Key: "status", Value: "*"},
	}}

	_ = getObject(query2)

	_ = postObject(key, "bla")
	_ = postObject(key2, "blue")
	_ = postObject(key3, "I have lots of questions. What's the meaning of life?")

	_ = getObject(key)
	_ = getObject(query)

	_ = deleteObject(key)

	_ = getObject(key2)
	_ = getObject(query2)

	// clean up
	_ = deleteObject(key2)
	_ = deleteObject(key3)
}

func main() {
	query := &pb.Key{Parts:[]*pb.Key_Part{
		{Key: "timestamp", Value: "*"},
		{Key: "id", Value:"*"},
		{Key: "category", Value:"*"},
		{Key: "status", Value: "*"},
	}}

	_ = getObject(query)
}
