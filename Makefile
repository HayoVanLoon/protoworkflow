
protoc:
	$(MAKE) -C proto protoc-go
	$(MAKE) -C contact_grpc protoc

build:
	$(MAKE) -C categorising_grpc build
	$(MAKE) -C contact_grpc build
	$(MAKE) -C messaging_grpc build
	$(MAKE) -C storage_grpc build


# Contact gRPC Server
build-contact:
	$(MAKE) -C contact_grpc build

run-contact:
	@$(MAKE) -C contact_grpc run

docker-run-contact:
	@$(MAKE) -C contact_grpc docker-run


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


# Storage gRPC Server
build-storage:
	@$(MAKE) -C storage_grpc build

run-storage:
	@$(MAKE) -C storage_grpc run

docker-run-storage:
	@$(MAKE) -C storage_grpc docker-run
