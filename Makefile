
protoc:
	$(MAKE) -C proto protoc-go
	$(MAKE) -C contact_grpc protoc

build:
	$(MAKE) -C contact_grpc build
	$(MAKE) -C categorising_grpc build

run-categorising:
	$(MAKE) -C categorising_grpc run