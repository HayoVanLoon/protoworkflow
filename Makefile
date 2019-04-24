
protoc:
	$(MAKE) -C proto protoc-go
	$(MAKE) -C contact_grpc protoc

build:
	$(MAKE) -C categorising_grpc build
	$(MAKE) -C contact_grpc build
	$(MAKE) -C messaging_grpc build
	$(MAKE) -C storage_grpc build

run-categorising:
	@$(MAKE) -C categorising_grpc run

docker-run-categorising:
	@$(MAKE) -C categorising_grpc docker-run

run-storage:
	@$(MAKE) -C storage_grpc run

docker-run-storage:
	@$(MAKE) -C storage_grpc docker-run

# Contact gRPC Server
build-contact:
	$(MAKE) -C contact_grpc build

run-contact:
	@$(MAKE) -C contact_grpc run

docker-run-contact:
	@$(MAKE) -C contact_grpc docker-run
