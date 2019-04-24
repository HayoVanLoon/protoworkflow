# Copyright 2019 Hayo van Loon
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

GOPATH ?= $(HOME)/go
GO_OUT := $(HOME)/go/src/github.com/HayoVanLoon/protoworkflow-genproto
GO_OUT2 := $(shell echo $(GO_OUT) | sed 's/\//\\\//g')

GO_PROTO_DIRS := $(shell find . -name '*.proto' | sed -E "s/^\.\/(.*)\/.*/$(GO_OUT2)\/\1/" | uniq)
PROTO_FILES := $(shell find . -name '*.proto')

clean-protoc:
	rm -rf $(GO_PROTO_DIRS)

protoc-go: clean-protoc
	mkdir -p $(GO_PROTO_DIRS)
	for FILE in $(PROTO_FILES); do \
		protoc \
			--go_out="plugins=grpc:$(GOPATH)/src" \
			-I$(PROTO_GOOGLE_APIS) \
			-I. \
			$$FILE; \
	done
