
PROJECT_NAME := protoworkflow
MODULE_NAME := sentiment

# VirtualEnv
VENV_DIR := venv
VENV := . $(VENV_DIR)/bin/activate

# Protocol Buffer Variables (Python)
ORGANISATION := protoworkflow
MODULE := contact
PROTO_VERSION := v1
PROTO_SELF := ../../proto/$(PROJECT_ORGANISATION)/$(MODULE)/$(PROTO_VERSION)
OUT := .


venv:
	rm -rf $(VENV_DIR)
	virtualenv -p python3.5 $(VENV_DIR)
	$(VENV); pip install -r requirements.txt

dev-env: venv

clean:
	rm -rf PLACEHOLDER_PREFIX.egg-info
	rm -rf dist
	rm -rf build

protoc: venv
ifndef PROTO_GOOGLE_APIS
	$(error PROTO_GOOGLE_APIS is not set, aborting)
else
	rm -rf *_pb2*.py
	$(VENV); python -m grpc_tools.protoc \
		--python_out=$(OUT) \
		--grpc_python_out=$(OUT) \
		-I$(PROTO_GOOGLE_APIS) \
		-I$(PROTO_SELF) \
		$(PROTO_SELF)/*.proto
endif

test: clean protoc
	$(VENV); pytest

run:
	$(VENV); python contact_server.py

test-call:
	$(VENV); python contact_client.py
