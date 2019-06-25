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

PROJECT_NAME := protoworkflow
MODULE_NAME := frontend

# Docker-related
IMAGE_NAME := $(PROJECT_NAME)_$(MODULE_NAME)_app
TAG := latest

# Protocol Buffer Variables
ORGANISATION := $(PROJECT_ORGANISATION)
MODULE := $(MODULE_NAME)
PROTO_VERSION := v1
PACKAGE_DIR := $(ORGANISATION)/$(MODULE)/$(PROTO_VERSION)


.PHONY:

protoc:
	@echo Go App, skipped

test:
	go test .

build:
	docker build -t $(IMAGE_NAME) .

run:
	go run frontend.go \
		-contact-host=$(shell minikube ip) \
		-contact-port=30000 \
		-messaging-host=$(shell minikube ip) \
		-messaging-port=30010

docker-run:
	docker run --network="host" $(IMAGE_NAME) \
		/usr/local/bin/app \
		-port=8080

push-gcr:
	docker tag $(IMAGE_NAME) gcr.io/$(PROJECT_ID)/$(IMAGE_NAME):$(TAG)
	docker push gcr.io/$(PROJECT_ID)/$(IMAGE_NAME)

#mocks:
#	mockgen -source $(GOPATH)/src/$(GENPROTO_REPO)/$(PACKAGE_DIR)/storage.pb.go
