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

CURRENT_VERSION := v1


protoc:
	@$(MAKE) -C v1 protoc

test:
	@$(MAKE) -C v1 test

build:
	@$(MAKE) -C $(CURRENT_VERSION) build

run:
	@$(MAKE) -C $(CURRENT_VERSION) run

docker-run:
	@$(MAKE) -C $(CURRENT_VERSION) docker-run

test-call:
	@$(MAKE) -C $(CURRENT_VERSION) test-call
