
PROJECT_NAME := protoworkflow
MODULE_NAME := sentiment

# Programs / Tools
GO := /usr/local/go/bin/go

# Protocol Buffer Variables (Go)
ORGANISATION := protoworkflow
MODULE := $(MODULE_NAME)
PROTO_VERSION := v1
PROTO_SELF := ../../proto/$(PROJECT_ORGANISATION)/$(MODULE)/$(PROTO_VERSION)
# Workaround for building go docker file
# TODO: check if workaround is still needed
GO_GENERATED := $(HOME)/go/src/github.com/HayoVanLoon/go-generated
OUT := $(GO_GENERATED)/$(MODULE_NAME)/v1

# Docker-related
IMAGE_NAME := $(PROJECT_NAME)_$(MODULE_NAME)_grpc
TAG := latest

TEST_ROOT := test
MOCK_TARGET := $(TEST_ROOT)/$(MODULE_NAME)/v1/$(MODULE_NAME)_mock.go


.PHONY:

clean:
	rm -rf $(OUT)

protoc: clean
	mkdir -p "$(OUT)"
	$(PROTOC_EXEC) --go_out="plugins=grpc:$(OUT)" \
		-I$(PROTO_GOOGLE_APIS) \
		-I$(PROTO_SELF) \
		$(PROTO_SELF)/*.proto

build: protoc
	$(DOCKER) build -t $(IMAGE_NAME) .

docker-run:
	$(DOCKER) run -p 8080:8080 $(IMAGE_NAME)

push-gcr:
	$(DOCKER) tag $(IMAGE_NAME) gcr.io/$(PROJECT_ID)/$(IMAGE_NAME):$(TAG)
	$(DOCKER) push gcr.io/$(PROJECT_ID)/$(IMAGE_NAME)

#
#mocks:
#	mkdir -p "$(TEST_ROOT)/$(ENVY_OUT)"
#	rm "$(MOCK_TARGET)"
#	mockgen -source envy/v1/envy.pb.go >> "$(MOCK_TARGET)"
#
#build:
#	$(go) build -i -o /dist/Envy_Server github.com/HayoVanLoon/envy/envy_server
#
#test-client: protoc
#	$(GO) build -i -o /tmp/___Envy_Client github.com/HayoVanLoon/envy/envy_client #gosetup
#	/tmp/___Envy_Client #gosetup
