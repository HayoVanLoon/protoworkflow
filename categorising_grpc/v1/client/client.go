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
	pb "github.com/HayoVanLoon/protoworkflow-genproto/bobsknobshop/categorising/v1"
	"google.golang.org/grpc"
	"log"
	"time"
)

const (
	defaultHost = "192.168.39.110"
	defaultPort = "8081"
)

func getConn(host, port string) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(host+":"+port, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("did not connect: %v", err)
	}
	return conn, nil
}

func getCategory(host, port, m string) error {
	r := &pb.GetCategoryRequest{Text: m}

	conn, err := getConn(host, port)
	defer func() {
		if err := conn.Close(); err != nil {
			log.Panicf("error closing connection: %v", err)
		}
	}()

	c := pb.NewCategorisingClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.GetCategory(ctx, r)

	log.Printf("%v\n", resp)

	return err
}

func main() {
	var host = flag.String("host", defaultHost, "categorising service host")
	var port = flag.String("port", defaultPort, "categorising service port")
	flag.Parse()

	fmt.Println(getCategory(*host, *port, "This does not please me."))
	fmt.Println(getCategory(*host, *port, "Everything is awesome."))
	fmt.Println(getCategory(*host, *port, "Ça je n'aime pas."))
	fmt.Println(getCategory(*host, *port, "I have a question about this product. Can I eat it?"))
}
