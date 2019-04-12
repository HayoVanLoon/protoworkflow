
PROJECT_NAME := protoworkflow
MODULE_NAME := messaging

# Protocol Buffer Variables
ORGANISATION := $(PROJECT_ORGANISATION)
MODULE := $(MODULE_NAME)
PROTO_VERSION := v1
PACKAGE_DIR := $(ORGANISATION)/$(MODULE)/$(PROTO_VERSION)
PROTO_ROOT := ../../proto
# Protocol Buffer Variables (Go)
GO_OUT := $(HOME)/go/src/github.com/HayoVanLoon/go-generated
GO_GENERATED := $(GO_OUT)/$(PACKAGE_DIR)

# Docker-related
IMAGE_NAME := $(PROJECT_NAME)_$(MODULE_NAME)_grpc
TAG := latest

TEST_ROOT := test
MOCK_TARGET := $(TEST_ROOT)/$(PACKAGE_DIR)/v1/$(MODULE_NAME)_mock.go


.PHONY:

clean: clean-protoc

clean-protoc:
	rm -rf $(GO_GENERATED)/*

protoc: clean-protoc
	mkdir -p "$(GO_GENERATED)"
	$(PROTOC_EXEC) \
		--go_out="plugins=grpc:$(GO_OUT)" \
		-I$(PROTO_GOOGLE_APIS) \
		-I$(PROTO_ROOT) \
		$(PROTO_ROOT)/$(PACKAGE_DIR)/*.proto

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
