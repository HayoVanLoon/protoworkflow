package main

import (
	pb "github.com/HayoVanLoon/protoworkflow-genproto/bobsknobshop/storage/v1"
	"reflect"
	"testing"
)

var (
	postMessage1 = &pb.CreateObjectRequest{
		Key: &pb.Key{
			Parts: []*pb.Key_Part{
				{Key: "foo", Value: "123"}, {Key: "bar", Value: "456"},
			},
			IndexedValues: []*pb.Key_Part{
				{Key: "colour", Value: "red"}, {Key: "shape", Value: "round"},
			},
		},
		Data: []byte("post message 1"),
	}
	postMessageResp1 = &pb.CreateObjectResponse{
		Name: "foo=123~bar=456",
	}
	postMessage2 = &pb.CreateObjectRequest{
		Key: &pb.Key{
			Parts: []*pb.Key_Part{
				{Key: "foo", Value: "123"}, {Key: "bar", Value: "654"},
			},
			IndexedValues: []*pb.Key_Part{
				{Key: "colour", Value: "red"}, {Key: "shape", Value: "square"},
			},
		},
		Data: []byte("post message 2"),
	}
	postMessage3 = &pb.CreateObjectRequest{
		Key: &pb.Key{
			Parts: []*pb.Key_Part{
				{Key: "foo", Value: "321"}, {Key: "bar", Value: "456"},
			},
			IndexedValues: []*pb.Key_Part{
				{Key: "colour", Value: "green"}, {Key: "shape", Value: "round"},
			},
		},
		Data: []byte("post message 3"),
	}
)

func createPostMessageResponse(name string) *pb.CreateObjectResponse {
	return &pb.CreateObjectResponse{
		Name: name,
	}
}

func createGetObjectRequest(key string) *pb.GetObjectRequest {
	return &pb.GetObjectRequest{
		Keys: []*pb.Key{{Name: key}},
	}
}

func createGetObjectQueryReq(keyParts []*pb.Key_Part) *pb.GetObjectRequest {
	return &pb.GetObjectRequest{
		Keys: []*pb.Key{{Parts: keyParts}},
	}
}

// quick & dirty demonstration test case, does not conclusively probe internal state
func TestServer_PostObject_GetObject(t *testing.T) {
	cases := []struct {
		s    *server
		req  *pb.CreateObjectRequest
		name string
		data string
	}{
		{newServer(), postMessage1, "foo=123~bar=456", "post message 1"},
		{newServer(), postMessage2, "foo=123~bar=654", "post message 2"},
		{newServer(), postMessage3, "foo=321~bar=456", "post message 3"},
	}
	for i, c := range cases {
		expected_resp := createPostMessageResponse(c.name)
		if resp, _ := c.s.CreateObject(nil, c.req); !reflect.DeepEqual(resp, expected_resp) {
			t.Errorf("case %v failed", i)
		}

		resp2, _ := c.s.GetObject(nil, createGetObjectRequest(c.name))
		if data := string(resp2.Data[0]); data != c.data {
			t.Errorf("case %v: expected %v, got %v", i, c.data, data)
		}
	}
}

func TestServer_Indexing(t *testing.T) {
	s := newServer()
	_, _ = s.CreateObject(nil, postMessage1)
	_, _ = s.CreateObject(nil, postMessage2)
	_, _ = s.CreateObject(nil, postMessage3)

	cases := []struct {
		query *pb.GetObjectRequest
		dataz []string
		fail  bool
	}{
		{
			createGetObjectQueryReq([]*pb.Key_Part{{Key: "colour", Value: "red"}}),
			[]string{"post message 1", "post message 2"},
			false,
		},
		{
			createGetObjectQueryReq([]*pb.Key_Part{{Key: "shape", Value: "round"}}),
			[]string{"post message 1", "post message 3"},
			false,
		},
		{
			createGetObjectQueryReq([]*pb.Key_Part{{Key: "shape", Value: "round"}}),
			[]string{"post message 1", "post message 3"},
			false,
		},
		{
			createGetObjectQueryReq([]*pb.Key_Part{}),
			nil,
			true,
		},
	}
	for i, c := range cases {
		resp, err := s.GetObject(nil, c.query)
		if c.fail {
			if err == nil {
				t.Errorf("case %v: expected failure", i)
			}
		} else if err != nil{
			t.Errorf("unexpected error: %v", err)
		} else {

			if len(resp.GetData()) != len(c.dataz) {
				t.Errorf("case %v: response count mismatch %v <> %v", i, len(resp.Data), len(c.dataz))
			}
			for _, left := range c.dataz {
				found := false
				for _, right := range s.data.items {
					found = found || left == string(right.data)
				}
				if !found {
					t.Errorf("case %v: missing in result: %v", i, left)
				}
			}
		}
	}
}
