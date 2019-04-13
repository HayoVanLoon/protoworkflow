
PROJECT_NAME := protoworkflow
MODULE_NAME := contact

# VirtualEnv
VENV_DIR := venv
VENV := . $(VENV_DIR)/bin/activate

# Protocol Buffer Variables
ORGANISATION := $(PROJECT_ORGANISATION)
MODULE := $(MODULE_NAME)
PROTO_VERSION := v1
PACKAGE_DIR := $(ORGANISATION)/$(MODULE)/$(PROTO_VERSION)
PROTO_ROOT := ../../proto
# Protocol Buffer Variables (Python)
PYTHON_OUT := .



venv:
	rm -rf $(VENV_DIR)
	virtualenv -p python3.5 $(VENV_DIR)
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
		$(PROTO_ROOT)/$(PACKAGE_DIR)/*.proto
endif

build:
	@echo Not implemented

test: clean protoc
	$(VENV); pytest

run:
	$(VENV); python contact_server.py

test-call:
	$(VENV); python contact_client.py
