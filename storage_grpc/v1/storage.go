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
	"crypto/sha1"
	"fmt"
	"github.com/HayoVanLoon/go-commons/sorted"
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

type keyVal struct {
	k, v string
}

func (kv keyVal) String() string {
	return fmt.Sprintf("keyVal{%s=%s}", kv.k, kv.v)
}

// Flattens a pb.Key so it can be used as a data map dkey
func toKey(k *pb.Key) dkey {
	if k.GetName() != "" {
		return dkey(k.GetName())
	}

	var ss []string
	for _, p := range k.Parts {
		ss = append(ss, p.Key+kvSep+p.Value)
	}
	return dkey(strings.Join(ss, sep))
}

// Turns indexed values into an index map
func toIdx(k *pb.Key, query bool) []keyVal {
	var result []keyVal
	for _, p := range k.Parts {
		result = append(result, keyVal{p.Key, p.Value})
		if !query {
			result = append(result, keyVal{p.Key, wildcard})
		}
	}
	for _, p := range k.IndexedValues {
		result = append(result, keyVal{p.Key, p.Value})
		if !query {
			result = append(result, keyVal{p.Key, wildcard})
		}
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
	idx  []keyVal
	data []byte
}

type dataMap struct {
	sync.RWMutex
	idxs  map[string]map[string]sorted.StringSet
	items map[dkey]item
}

type server struct {
	data dataMap
}

func newServer() *server {
	dataMap := dataMap{
		idxs: make(map[string]map[string]sorted.StringSet),
		items: make(map[dkey]item),
	}
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

func (s *server) getKeys(query []keyVal) []string {
	s.data.RLock()
	defer s.data.RUnlock()

	var left []string
	for _, queryKv := range query {
		if vs, ok := s.data.idxs[queryKv.k]; ok {
			if ks, ok := vs[queryKv.v]; ok {
				if left == nil {
					left = ks.Slice()
				} else {
					i, j := 0, 0
					right := ks.Slice()
					var intersect []string
					for ; i < len(left) && j < len(right); {
						if left[i] > right[i] {
							j += 1
						} else if left[j] < right[i] {
							i += 1
						} else {
							intersect = append(intersect, left[j])
							i += 1
							j += 1
						}
					}
					left = intersect
				}
			}
		}
	}

	return left
}

func (s *server) putData(key dkey, idx []keyVal, d []byte) (dkey, error) {
	s.data.Lock()
	defer s.data.Unlock()

	if _, ex := s.data.items[key]; ex {
		m := fmt.Sprintf("already have message with key %s", key)
		log.Print(m)
		return "", fmt.Errorf(m)
	}

	it := item{idx: idx, data: d}
	s.data.items[key] = it

	s.addToIdxs(it.idx, key)

	return key, nil
}

// MUST be under mutex!
func (s *server) addToIdxs(idx []keyVal, key dkey) {
	for _, kv := range idx {
		if vs, ok := s.data.idxs[kv.k]; ok {
			if ks, ok := vs[kv.v]; ok {
				ks = ks.Add(string(key))
				s.data.idxs[kv.k][kv.v]= ks
			} else {
				vs[kv.v] = sorted.NewStringSet().Add(string(key))
				s.data.idxs[kv.k] = vs
			}
		} else {
			newVs := sorted.NewStringSet().Add(string(key))
			s.data.idxs[kv.k] = map[string]sorted.StringSet{kv.v: newVs}
		}
		log.Printf("DEBUG: added key %v to index (%v, %v)", key, kv.k, kv.v)
	}
}

// MUST be under mutex!
func (s *server) deleteFromIdxs(idx []keyVal, key dkey) {
	for _, kv := range idx {
		if vs, ok := s.data.idxs[kv.k]; ok {
			if ks, ok := vs[kv.v]; ok {
				ks.Remove(string(key))
				s.data.idxs[kv.k][kv.v] = ks
				log.Printf("DEBUG: deleted key %v from index (%v, %v)", key, kv.k, kv.v)
			}
		}
	}
}

func (s *server) deleteData(key dkey) {
	s.data.Lock()
	defer s.data.Unlock()
	if it, ok := s.data.items[key]; ok {
		s.deleteFromIdxs(it.idx, key)
		delete(s.data.items, key)
	}
}

func getEtag(data []byte) string {
	h := sha1.New()
	bs := h.Sum(data)
	return fmt.Sprintf("%x", bs)
}

func (s *server) mutateData(oldKey, newKey *pb.Key, oldEtag string, newData []byte) (bool, string) {
	s.data.Lock()
	defer s.data.Unlock()

	key := toKey(oldKey)
	if it, ok := s.data.items[key]; ok {
		if curEtag := getEtag(it.data); curEtag == oldEtag {
			newIt := item{idx: toIdx(newKey, false), data: newData}
			s.data.items[key] = newIt
			s.deleteFromIdxs(toIdx(oldKey, false), key)
			s.addToIdxs(toIdx(newKey, false), key)
			return true, ""
		} else {
			return false, getEtag(it.data)
		}
	} else {
		return false, ""
	}
}

func (s *server) CreateObject(_ context.Context, req *pb.CreateObjectRequest) (*pb.CreateObjectResponse, error) {
	key, err := s.putData(toKey(req.Key), toIdx(req.Key, false), req.Data)
	if err != nil{
		return nil, fmt.Errorf("could not store %s", key)
	}
	log.Printf("DEBUG: stored %s", key)
	return &pb.CreateObjectResponse{Name: string(key)}, nil
}

func (s *server) GetObject(_ context.Context, req *pb.GetObjectRequest) (*pb.GetObjectResponse, error) {
	// use intermediate map to prevent duplicates in result
	result := make(map[dkey][]byte)

	for _, k := range req.Keys {
		asKey := toKey(k)
		if d, ok := s.getData(asKey); ok {
			// if dkey matches completely, there are no wildcards
			result[asKey] = d
		} else if len(k.Parts) > 0 {
			query := toIdx(k, true)
			for _, k2 := range s.getKeys(query) {
				if d, ok = s.getData(dkey(k2)); ok {
					result[dkey(k2)] = d
				}
			}
		} else {
			m := fmt.Sprintf("empty query")
			log.Print(m)
			return nil, fmt.Errorf(m)
		}
	}

	var ks []*pb.Key
	var ds [][]byte
	for k, d := range result {
		ks = append(ks, toPb(k))
		ds = append(ds, d)
	}
	log.Printf("DEBUG: returned %v objects for query", len(result))
	return &pb.GetObjectResponse{Keys: ks, Data: ds}, nil
}

func (s *server) DeleteObject(_ context.Context, req *pb.DeleteObjectRequest) (*empty.Empty, error) {
	for _, k := range req.Keys {
		key := toKey(k)
		if _, ok := s.getData(key); ok {
			s.deleteData(key)
			log.Printf("DEBUG: deleted %s", key)
		}
	}
	return &empty.Empty{}, nil
}

func (s *server) MutateObject(ctx context.Context, req *pb.MutateObjectRequest) (*pb.MutateObjectResponse, error) {
	ok, etag := s.mutateData(req.GetOldKey(), req.GetNewKey(), req.GetOldEtag(), req.GetNewData())
	if ok {
		log.Printf("DEBUG: updated %s", toKey(req.GetOldKey()))
	} else {
		log.Printf("INFO: could not update %s", toKey(req.GetOldKey()))
	}
	return &pb.MutateObjectResponse{Success: ok, NewEtag: etag}, nil
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
