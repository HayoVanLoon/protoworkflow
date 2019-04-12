
protoc:
	$(MAKE) -C contact_grpc protoc
	$(MAKE) -C messaging_grpc protoc
	$(MAKE) -C sentiment_grpc protoc

build:
	$(MAKE) -C contact_grpc build
	$(MAKE) -C sentiment_grpc build
