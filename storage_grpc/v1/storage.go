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
	"bytes"
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

// alias for readability
type dkey string

// Flattens a pb.Key so it can be used as a data map dkey
func toKey(k *pb.Key) dkey {
	var ss []string
	for _, p := range k.Parts {
		ss = append(ss, p.Key+kvSep+p.Value)
	}
	return dkey(strings.Join(ss, sep))
}

// Turns indexed values into an index map
func toIdx(k *pb.Key) map[string]string {
	result := make(map[string]string)
	for _, p := range k.Parts {
		result[p.Key] = p.Value
	}
	for _, p := range k.IndexedValues {
		result[p.Key] = p.Value
	}
	return result
}

// Reconstructs a pb.Key from a data map dkey string
func toPb(k dkey) *pb.Key {
	ss := strings.Split(string(k), sep)
	var ps []*pb.Key_Part
	for _, kv := range ss {
		ks := strings.Split(kv, kvSep)
		ps = append(ps, &pb.Key_Part{Key: ks[0], Value: ks[1]})
	}
	return &pb.Key{Parts: ps}
}

type item struct {
	idxs map[string]string
	data []byte
}

type dataMap struct {
	sync.RWMutex
	idxs  map[string]map[string][]dkey
	items map[dkey]item
}

type server struct {
	data dataMap
}

func newServer() *server {
	dataMap := dataMap{idxs: make(map[string]map[string][]dkey), items: make(map[dkey]item)}
	return &server{dataMap}
}

func (s *server) getData(key dkey) ([]byte, bool) {
	s.data.RLock()
	defer s.data.RUnlock()
	if it, ok := s.data.items[key]; ok {
		return it.data, ok
	}
	return nil, false
}

func (s *server) getKeys(query map[string]string) []dkey {
	s.data.RLock()
	defer s.data.RUnlock()

	var result []dkey
	for k, v := range query {
		if vs, ok := s.data.idxs[k]; ok {
			if ks, ok := vs[v]; ok {
				if result == nil {
					result = ks
				} else {
					i, j := 0, 0
					var newR []dkey
					for ; i < len(ks) && j < len(result); {
						if result[j] > ks[i] {
							i += 1
						} else if result[j] < ks[i] {
							j += 1
						} else {
							newR = append(newR, result[j])
							i += 1
							j += 1
						}
					}
					result = newR
				}
			}
		}
	}

	return result
}

func (s *server) putData(key dkey, idx map[string]string, d []byte) bool {
	s.data.Lock()
	defer s.data.Unlock()

	if _, ex := s.data.items[key]; ex {
		return false
	}

	it := item{idxs: idx, data: d}
	s.data.items[key] = it

	updateIndexes(s.data.idxs, it, key)

	return true
}

// MUST be under mutex!
func updateIndexes(idxs map[string]map[string][]dkey, it item, key dkey) {
	for k, v := range it.idxs {
		if vs, ok := idxs[k]; ok {
			if ks, ok := vs[v]; ok {
				// insert dkey into a sorted array
				for i, k2 := range ks {
					if key < k2 {
						ks := append(ks, "")
						copy(ks[i+1:], ks[i:])
						ks[i] = key
						break
					}
				}
			} else {
				vs[v] = []dkey{key}
			}
			vs[wildcard] = append(vs[wildcard], key)
		} else {
			idxs[k] = map[string][]dkey{v: {key}}
		}
	}
}

// MUST be under mutex!
func deleteFromIdx(idxs map[string]map[string][]dkey, it item, key dkey) {
	for k, v := range it.idxs {
		if vs, ok := idxs[k]; ok {
			if ks, ok := vs[v]; ok {
				for i, k2 := range ks {
					if k2 == key {
						ks = append(ks[:i], ks[i+1:]...)
						return
					}
				}
			}
		}
	}
}

func (s *server) deleteData(key dkey) {
	s.data.Lock()
	defer s.data.Unlock()
	if it, ok := s.data.items[key]; ok {
		deleteFromIdx(s.data.idxs, it, key)
		delete(s.data.items, key)
	}
}

func (s *server) mutateData(oldKey, newKey *pb.Key, oldData, newData []byte) (bool, []byte) {
	s.data.Lock()
	defer s.data.Unlock()

	key := toKey(oldKey)
	if it, ok := s.data.items[key]; ok {
		if bytes.Equal(it.data, oldData) {
			newIt := item{idxs: it.idxs, data: newData}
			s.data.items[key] = newIt
			// TODO: remove old indices
			// TODO: update indices
			return true, nil
		} else {
			return false, it.data
		}
	} else {
		return false, nil
	}
}

func (s *server) PostObject(_ context.Context, req *pb.PostObjectRequest) (*pb.PostObjectResponse, error) {
	ok := s.putData(toKey(req.Key), toIdx(req.Key), req.Data)
	return &pb.PostObjectResponse{Success: ok}, nil
}

func (s *server) GetObject(_ context.Context, req *pb.GetObjectRequest) (*pb.GetObjectResponse, error) {
	// use intermediate map to prevent duplicates in result
	result := make(map[dkey][]byte)

	for _, k := range req.Keys {
		asKey := toKey(k)
		if d, ok := s.getData(asKey); ok {
			// if dkey matches completely, there are no wildcards
			result[asKey] = d
		} else {
			query := toIdx(k)
			for _, k2 := range s.getKeys(query) {
				if d, ok = s.getData(k2); ok {
					result[k2] = d
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
		}
	}
	return &empty.Empty{}, nil
}

func (s *server) MutateObject(ctx context.Context, req *pb.MutateObjectRequest) (*pb.MutateObjectResponse, error) {
	ok, current := s.mutateData(req.GetOldKey(), req.GetNewKey(), req.GetOldData(), req.GetNewData())
	return &pb.MutateObjectResponse{Success: ok, Current: current}, nil
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
