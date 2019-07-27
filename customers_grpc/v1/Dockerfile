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

FROM alpine

RUN apk --update --no-cache add ca-certificates openssl

# Referenced from https://github.com/grpc/grpc/blob/master/tools/dockerfile/test/python_alpine_x64/Dockerfile:
RUN apk --no-cache add \
    autoconf automake bzip2 build-base cmake ccache gcc libtool linux-headers \
    make perl strace unzip wget zip

RUN apk --no-cache add python3 python3-dev py3-pip

WORKDIR /app

RUN /usr/bin/pip3 install protobuf
# really slow:
RUN /usr/bin/pip3 install grpcio
# really slow:
RUN /usr/bin/pip3 install grpcio-tools

COPY requirements/docker.txt ./requirements.txt
RUN /usr/bin/pip3 install -r requirements.txt

COPY bobsknobshop/ ./bobsknobshop/
COPY server.py .

CMD ["/usr/bin/python3", "server.py"]
