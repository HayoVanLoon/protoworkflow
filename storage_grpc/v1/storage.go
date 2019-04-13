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
	pb "github.com/HayoVanLoon/protoworkflow-genproto/bobsknobshop/storage/v1"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"strings"
)

const (
	port              = "8080"
	wildcard = "*"
	sep = "~"
	kvSep = "="
)

func toLocal(k *pb.Key) string {
	var ss []string
	for _, p := range k.Parts {
		ss = append(ss, p.Key + kvSep + p.Value)
	}
	return strings.Join(ss, sep)
}

func toPb(s string) pb.Key {
	ss := strings.Split(s, sep)
	var ps []*pb.Key_Part
	for _, kv := range ss {
		ks := strings.Split(kv, kvSep)
		ps = append(ps, &pb.Key_Part{Key: ks[0], Value: ks[1]})
	}
	return pb.Key{Parts: ps}
}

type server struct {
	ds map[string][]byte
}

func (s *server) PostObject(_ context.Context, req *pb.PostObjectRequest) (*pb.PostObjectResponse, error) {
	s.ds[toLocal(req.Key)] = req.Data
	return &pb.PostObjectResponse{}, nil
}

func (s *server) GetObject(_ context.Context, req *pb.GetObjectRequest) (*pb.GetObjectResponse, error) {
	var ks []*pb.Key
	var ds [][]byte
	for _, sk := range req.Keys {
		if d, ok := s.ds[toLocal(sk)]; ok {
			ks = append(ks, sk)
			ds = append(ds, d)
		} else {
			for ss, d := range s.ds {
				kvs := toPb(ss)
				ok = len(sk.Parts) == len(kvs.Parts)
				for i := 0; ok && i < len(kvs.Parts); i += 1 {
					left, right := sk.Parts[i], kvs.Parts[i]
					ok = ok && left.Key == right.Key && left.Value == right.Value || left.Value == wildcard
				}
				if ok {
					ks = append(ks, sk)
					ds = append(ds, d)
				}
			}
		}
	}
	return &pb.GetObjectResponse{Keys: ks, Data: ds}, nil
}

func (s *server) DeleteObject(_ context.Context, req *pb.DeleteObjectRequest) (*empty.Empty, error) {
	for _, k := range req.Keys {
		kvs := toLocal(k)
		if _, f := s.ds[kvs]; f {
			delete(s.ds, kvs)
		} else {
			// TODO: return not-found error
		}
	}
	return &empty.Empty{}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterStorageServer(s, &server{make(map[string][]byte)})

	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
