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
MODULE_NAME := customers

# Docker-related
IMAGE_NAME := $(PROJECT_NAME)_$(MODULE_NAME)_grpc
TAG := latest

# VirtualEnv
PY != which python3.6
VENV_DIR := venv
VENV := . $(VENV_DIR)/bin/activate

# Protocol Buffer Variables
ORGANISATION := $(PROJECT_ORGANISATION)
MODULE := $(MODULE_NAME)
PROTO_VERSION := v1
PACKAGE_DIR := $(ORGANISATION)/$(MODULE)/$(PROTO_VERSION)
PROTO_ROOT := ../../proto
# Protocol Buffer Variables (module)
PROTO_DEPS := $(PROTO_ROOT)/$(ORGANISATION)/messaging/$(PROTO_VERSION)/*.proto
PYTHON_OUT := .


venv:
	rm -rf $(VENV_DIR)
	virtualenv -p $(PY) $(VENV_DIR)
	$(VENV); pip install -r requirements.txt

dev-env: venv

clean: clean-protoc
	rm -rf PLACEHOLDER_PREFIX.egg-info
	rm -rf dist
	rm -rf build

clean-protoc:
	rm -rf $(PYTHON_OUT)/*_pb2*.py

protoc: venv clean-protoc
ifndef PROTO_GOOGLE_APIS
	$(error PROTO_GOOGLE_APIS is not set, aborting)
else
	$(VENV); python -m grpc_tools.protoc \
		--python_out=$(PYTHON_OUT) \
		--grpc_python_out=$(PYTHON_OUT) \
		-I$(PROTO_GOOGLE_APIS) \
		-I$(PROTO_ROOT) \
		$(PROTO_ROOT)/$(PACKAGE_DIR)/*.proto \
		$(PROTO_DEPS)
endif

test: clean protoc
#	$(VENV); pytest
	@echo TODO

build: build-python-grpc
	docker build -f Dockerfile-partial-build -t $(IMAGE_NAME) .

build-python-grpc:
	docker build -f Dockerfile-python-grpc -t python-grpc .

run:
	$(VENV); python server.py

docker-run:
	docker run --network="host" $(IMAGE_NAME) \
		/usr/bin/python3 server.py \
		--port=8083 \
		--messaging_host=localhost \
		--messaging_port=8082

test-minikube:
	$(VENV); python client.py --host=$(shell minikube ip) --port=30000
