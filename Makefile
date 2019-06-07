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

protoc:
	$(MAKE) -C proto protoc-go
	$(MAKE) -C contact_grpc protoc

test: protoc
	$(MAKE) -C categorising_grpc test
	$(MAKE) -C contact_grpc test
	$(MAKE) -C messaging_grpc test
	$(MAKE) -C storage_grpc test

build: test build-categorising build-contact build-messaging build-storage


# Contact gRPC Server
build-contact:
	$(MAKE) -C contact_grpc build

run-contact:
	@$(MAKE) -C contact_grpc run

docker-run-contact:
	@$(MAKE) -C contact_grpc docker-run

test-minikube-contact:
	@$(MAKE) -C contact_grpc test-minikube

# Categorising gRPC Server
build-categorising:
	$(MAKE) -C categorising_grpc build

run-categorising:
	@$(MAKE) -C categorising_grpc run

docker-run-categorising:
	@$(MAKE) -C categorising_grpc docker-run


# Messaging gRPC Server
build-messaging:
	$(MAKE) -C messaging_grpc build

run-messaging:
	@$(MAKE) -C messaging_grpc run

docker-run-messaging:
	@$(MAKE) -C messaging_grpc docker-run

test-minikube-messaging:
	@$(MAKE) -C messaging_grpc test-minikube

# Storage gRPC Server
build-storage:
	@$(MAKE) -C storage_grpc build

run-storage:
	@$(MAKE) -C storage_grpc run

docker-run-storage:
	@$(MAKE) -C storage_grpc docker-run
