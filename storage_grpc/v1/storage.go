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
	"sync"
)

const (
	port     = "8080"
	wildcard = "*"
	sep      = "~"
	kvSep    = "="
)

type key string

// Flattens a pb.Key so it can be used as a data map key
func toKey(k *pb.Key) key {
	var ss []string
	for _, p := range k.Parts {
		ss = append(ss, p.Key+kvSep+p.Value)
	}
	return key(strings.Join(ss, sep))
}

// Reconstructs a pb.Key from a data map key string
func toPb(k key) *pb.Key {
	ss := strings.Split(string(k), sep)
	var ps []*pb.Key_Part
	for _, kv := range ss {
		ks := strings.Split(kv, kvSep)
		ps = append(ps, &pb.Key_Part{Key: ks[0], Value: ks[1]})
	}
	return &pb.Key{Parts: ps}
}

type dataMap struct {
	sync.RWMutex
	ds map[key][]byte
}

type server struct {
	data dataMap
}

func newServer() *server {
	return &server{dataMap{ds: make(map[key][]byte)}}
}

func (s *server) getData(key key) ([]byte, bool) {
	s.data.RLock()
	defer s.data.RUnlock()
	d, k := s.data.ds[key]
	return d, k
}

func (s *server) getKeys() []key {
	s.data.RLock()
	defer s.data.RUnlock()
	ks := make([]key, 0, len(s.data.ds))
	for k := range s.data.ds {
		ks = append(ks, k)
	}
	return ks
}

func (s *server) putData(key key, d []byte) {
	s.data.Lock()
	defer s.data.Unlock()
	s.data.ds[key] = d
}

func (s *server) deleteData(key key) {
	s.data.Lock()
	defer s.data.Unlock()
	delete(s.data.ds, key)
}

func (s *server) PostObject(_ context.Context, req *pb.PostObjectRequest) (*pb.PostObjectResponse, error) {
	s.putData(toKey(req.Key), req.Data)
	return &pb.PostObjectResponse{}, nil
}

func (s *server) GetObject(_ context.Context, req *pb.GetObjectRequest) (*pb.GetObjectResponse, error) {
	// use intermediate map to prevent duplicates in result
	result := make(map[key][]byte)

	for _, query := range req.Keys {
		asKey := toKey(query)
		if d, ok := s.getData(asKey); ok {
			// if key matches completely, there are no wildcards
			result[asKey] = d
		} else {
			// no fancy indexes, just a complete table scan
			for _, k := range s.getKeys() {
				keyPb := toPb(k)
				ok = len(query.Parts) == len(keyPb.Parts)
				for i := 0; ok && i < len(keyPb.Parts); i += 1 {
					left, right := query.Parts[i], keyPb.Parts[i]
					ok = ok && left.Key == right.Key && left.Value == right.Value || left.Value == wildcard
				}
				if d, lastCheck := s.getData(k); ok && lastCheck {
					result[k] = d
				}
			}
		}
	}

	var ks []*pb.Key
	var ds [][]byte
	for k, d := range result {
		ks = append(ks, toPb(k))
		ds = append(ds, d)
	}
	return &pb.GetObjectResponse{Keys: ks, Data: ds}, nil
}

func (s *server) DeleteObject(_ context.Context, req *pb.DeleteObjectRequest) (*empty.Empty, error) {
	for _, k := range req.Keys {
		kvs := toKey(k)
		if _, ok := s.getData(kvs); ok {
			s.deleteData(kvs)
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
	pb.RegisterStorageServer(s, newServer())

	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
