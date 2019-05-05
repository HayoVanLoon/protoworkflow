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
MODULE_NAME := measuring

# Docker-related
IMAGE_NAME := $(PROJECT_NAME)_$(MODULE_NAME)_grpc
TAG := latest

# Protocol Buffer Variables
ORGANISATION := $(PROJECT_ORGANISATION)
MODULE := $(MODULE_NAME)
PROTO_VERSION := v1
PACKAGE_DIR := $(ORGANISATION)/$(MODULE)/$(PROTO_VERSION)
PROTO_ROOT := ../../proto
# Protocol Buffer Variables (module)
JAVA_OUT := target/generated-sources/protobuf/java

.PHONY:

clean:
	mvn clean
	rm -rf target

protoc: clean
	mvn protobuf:compile

compile:
	mvn compile

test:
	mvn verify

build:
	docker build -t $(IMAGE_NAME) .

run:
	mvn exec:java -Dexec.mainClass=gl.bobsknobshop.measuring.MeasuringServer

test-call:
	mvn exec:java -Dexec.mainClass=gl.bobsknobshop.measuring.TestClient

docker-run:
	@echo TODO

test-minikube:
	@echo TODO

push-gcr:
	docker tag $(IMAGE_NAME) gcr.io/$(PROJECT_ID)/$(IMAGE_NAME):$(TAG)
	docker push gcr.io/$(PROJECT_ID)/$(IMAGE_NAME)
