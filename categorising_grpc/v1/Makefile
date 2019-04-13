
PROJECT_NAME := protoworkflow
MODULE_NAME := categorising

# Docker-related
IMAGE_NAME := $(PROJECT_NAME)_$(MODULE_NAME)_grpc
TAG := latest

# Protocol Buffer Variables
ORGANISATION := $(PROJECT_ORGANISATION)
MODULE := $(MODULE_NAME)
PROTO_VERSION := v1
PACKAGE_DIR := $(ORGANISATION)/$(MODULE)/$(PROTO_VERSION)

TEST_ROOT := test
MOCK_TARGET := $(TEST_ROOT)/$(PACKAGE_DIR)/$(MODULE_NAME)_mock.go


.PHONY:

protoc:
	@echo Go App, skipped

build:
	$(DOCKER) build -t $(IMAGE_NAME) .

run:
	go run categorising.go

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
