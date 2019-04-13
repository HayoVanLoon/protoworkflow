
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
