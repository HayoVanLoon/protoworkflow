
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
	docker build -t $(IMAGE_NAME) .

run:
	go run categorising.go

docker-run:
	docker run --network="host" $(IMAGE_NAME) \
		/usr/local/bin/app \
		-port=8081

push-gcr:
	docker tag $(IMAGE_NAME) gcr.io/$(PROJECT_ID)/$(IMAGE_NAME):$(TAG)
	docker push gcr.io/$(PROJECT_ID)/$(IMAGE_NAME)

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
