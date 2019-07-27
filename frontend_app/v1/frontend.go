/*
 * Copyright 2019 Hayo van Loon
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
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
	"github.com/HayoVanLoon/protoworkflow-genproto/bobsknobshop/contact/v1"
	"github.com/HayoVanLoon/protoworkflow-genproto/bobsknobshop/messaging/v1"
	"github.com/HayoVanLoon/protoworkflow-genproto/bobsknobshop/storage/v1"
	"github.com/golang/protobuf/jsonpb"
	"google.golang.org/grpc"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	contactService   = "contact-service"
	messagingService = "messaging-service"
	storageService   = "storage-service"
	defaultPort      = 8080
)

type static struct {
	contentType string
	data        []byte
}

var statics = mapStatics("static", "")

func newStatic(file string) static {
	data, _ := ioutil.ReadFile(file)
	xs := strings.Split(file, ".")
	ext := xs[len(xs)-1]
	if ext == "css" {
		return static{"text/css", data}
	} else if ext == "js" {
		return static{"application/javascript", data}
	} else {
		return static{"text/html", []byte("")}
	}
}

func mapStatics(file string, path string) map[string]static {
	m := map[string]static{}

	var fp string
	if path == "" {
		fp = file
	} else {
		fp = filepath.Join(path, file)
	}

	fi, _ := os.Stat(fp)
	if fi.IsDir() {
		dir, _ := os.Open(fp)
		fis, _ := dir.Readdir(0)
		for _, fi2 := range fis {
			m2 := mapStatics(fi2.Name(), fp)
			for n, st := range m2 {
				m[n] = st
			}
		}
	} else {
		m[fp] = newStatic(fp)
	}

	return m
}

func handler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("index.html"))
	err := t.Execute(w, t)
	if err != nil {
		log.Print(err)
	}
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[1:]
	if st, ok := statics[path]; ok {
		w.Header()["Content-Type"] = []string{st.contentType}
		_, err := fmt.Fprint(w, string(st.data))
		if err != nil {
			log.Printf("%v\n", err)
			return
		}
	} else {
		log.Print(r.URL.Path)
		w.WriteHeader(404)
	}
}

// Establishes a connection to a service
func getConn(host, port string) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(host+":"+port, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("did not connect: %v", err)
	}
	return conn, nil
}

// Provides a connection-closing function
func closeConnFn(conn *grpc.ClientConn) func() {
	return func() {
		if err := conn.Close(); err != nil {
			log.Printf("WARN: error closing connection: %v", err)
		}
	}
}

// Provides a contact client
func getContactClient(host, port string) (contact.ContactClient, func(), error) {
	conn, err := getConn(host, port)
	if err != nil {
		log.Printf("ERROR: could not open connection to %s:%s", host, port)
		return nil, nil, err
	}
	return contact.NewContactClient(conn), closeConnFn(conn), err
}

// Provides a messaging client
func getMessagingClient(host, port string) (messaging.MessagingClient, func(), error) {
	conn, err := getConn(host, port)
	if err != nil {
		log.Printf("ERROR: could not open connection to %s:%s", host, port)
		return nil, nil, err
	}
	return messaging.NewMessagingClient(conn), closeConnFn(conn), err
}

// Provides a storage client
func getStorageClient(host, port string) (storage.StorageClient, func(), error) {
	conn, err := getConn(host, port)
	if err != nil {
		log.Printf("ERROR: could not open connection to %s:%s", host, port)
		return nil, nil, err
	}
	return storage.NewStorageClient(conn), closeConnFn(conn), err
}

func contactHandlerFn(host, port string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			_ = r.ParseForm()
			msgs, ok := r.Form["message"]
			if !ok || len(msgs) <= 0 {
				w.WriteHeader(400)
				return
			}
			ss, ok := r.Form["sender"]
			if !ok || len(ss) <= 0 {
				w.WriteHeader(400)
				return
			}

			m := &contact.CreateMessageRequest{
				Message: msgs[0],
				Sender:  &contact.Sender{Name: ss[0]},
			}
			if pi, ok := r.Form["product-id"]; ok && len(pi) > 0 {
				m.ProductId = pi[0]
			}

			c, closeConn, err := getContactClient(host, port)
			if err != nil {
				log.Printf("%v", err)
				return
			}
			defer closeConn()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			_, err = c.CreateMessage(ctx, m)
			if err != nil {
				log.Printf("%v", err)
				w.WriteHeader(500)
				return
			}

			_, _ = fmt.Fprintf(w, "posted")
		} else {
			w.WriteHeader(405)
		}
	}
}

func messagesHandlerFn(host, port string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		c, closeConn, err := getMessagingClient(host, port)
		if err != nil {
			log.Printf("error creating messaging client %v", err)
			w.WriteHeader(500)
			return
		}
		defer closeConn()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		m := &messaging.SearchMessagesRequest{
			Status: []messaging.Status{messaging.Status_TO_DO},
		}

		resp, err := c.SearchMessages(ctx, m)
		if err != nil {
			log.Printf("error calling SearchMessages %v", err)
			w.WriteHeader(500)
			return
		}

		mars := jsonpb.Marshaler{}
		w.Header()["Content-Type"] = []string{"application/json"}
		_ = mars.Marshal(w, resp)
	}
}

func getQuestionHandlerFn(host, port string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		c, closeConn, err := getMessagingClient(host, port)
		if err != nil {
			log.Printf("error creating messaging client %v", err)
			w.WriteHeader(500)
			return
		}
		defer closeConn()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		m := &messaging.GetQuestionRequest{}

		resp, err := c.GetQuestion(ctx, m)
		if err != nil {
			log.Printf("error calling GetQuestion %v", err)
			w.WriteHeader(500)
			return
		}

		mars := jsonpb.Marshaler{}
		w.Header()["Content-Type"] = []string{"application/json"}
		_ = mars.Marshal(w, resp)
	}
}

func getComplaintHandlerFn(host, port string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		c, closeConn, err := getMessagingClient(host, port)
		if err != nil {
			log.Printf("error creating messaging client %v", err)
			w.WriteHeader(500)
			return
		}
		defer closeConn()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		m := &messaging.GetComplaintRequest{}

		resp, err := c.GetComplaint(ctx, m)
		if err != nil {
			log.Printf("error calling GetComplaint %v", err)
			w.WriteHeader(500)
			return
		}

		mars := jsonpb.Marshaler{}
		w.Header()["Content-Type"] = []string{"application/json"}
		_ = mars.Marshal(w, resp)
	}
}

func getFeedbackHandlerFn(host, port string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		c, closeConn, err := getMessagingClient(host, port)
		if err != nil {
			log.Printf("error creating messaging client %v", err)
			w.WriteHeader(500)
			return
		}
		defer closeConn()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		m := &messaging.GetFeedbackRequest{}

		resp, err := c.GetFeedback(ctx, m)
		if err != nil {
			log.Printf("error calling GetFeedback %v", err)
			w.WriteHeader(500)
			return
		}

		mars := jsonpb.Marshaler{}
		w.Header()["Content-Type"] = []string{"application/json"}
		_ = mars.Marshal(w, resp)
	}
}

func getStorageStatsHandlerFn(host, port string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		c, closeConn, err := getStorageClient(host, port)
		if err != nil {
			log.Printf("error creating messaging client %v", err)
			w.WriteHeader(500)
			return
		}
		defer closeConn()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		m := &storage.GetStatsRequest{}

		resp, err := c.GetStats(ctx, m)
		if err != nil {
			log.Printf("error calling GetStorageStats %v", err)
			w.WriteHeader(500)
			return
		}

		mars := jsonpb.Marshaler{}
		w.Header()["Content-Type"] = []string{"application/json"}
		_ = mars.Marshal(w, resp)
	}
}

func main() {
	var port = flag.Int("port", defaultPort, "port to listen on")
	var contactHost = flag.String("contact-host", contactService, "contact service")
	var contactPort = flag.Int("contact-port", defaultPort, "contact service port")
	var messagingHost = flag.String("messaging-host", messagingService, "messaging service")
	var messagingPort = flag.Int("messaging-port", defaultPort, "messaging service port")
	var storageHost = flag.String("storage-host", storageService, "storage service")
	var storagePort = flag.Int("storage-port", defaultPort, "storage service port")
	flag.Parse()

	http.HandleFunc("/", handler)
	http.HandleFunc("/contact", contactHandlerFn(*contactHost, strconv.Itoa(*contactPort)))
	// http.HandleFunc("/messages", messagesHandlerFn(*messagingHost, *messagingPort))
	http.HandleFunc("/question", getQuestionHandlerFn(*messagingHost, strconv.Itoa(*messagingPort)))
	http.HandleFunc("/complaint", getComplaintHandlerFn(*messagingHost, strconv.Itoa(*messagingPort)))
	http.HandleFunc("/feedback", getFeedbackHandlerFn(*messagingHost, strconv.Itoa(*messagingPort)))
	http.HandleFunc("/storage-stats", getStorageStatsHandlerFn(*storageHost, strconv.Itoa(*storagePort)))
	http.HandleFunc("/static/", staticHandler)

	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(*port), nil))
}
