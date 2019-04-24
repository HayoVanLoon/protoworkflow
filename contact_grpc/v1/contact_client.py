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

import argparse
import grpc
from bobsknobshop.contact.v1 import contact_pb2
from bobsknobshop.contact.v1 import contact_pb2_grpc
from google.protobuf import json_format


DEFAULT_HOST = 'localhost'
DEFAULT_PORT = 8080
DEFAULT_TIMEOUT = 10


def main(params):
    target = params['host'] + ':' + params['port']
    with grpc.insecure_channel(target) as channel:
        stub = contact_pb2_grpc.ContactStub(channel)

        message = contact_pb2.PostMessageRequest()
        message.message = 'This does not please me.'

        resp = stub.PostMessage(message, timeout=DEFAULT_TIMEOUT)
        print(json_format.MessageToJson(resp))


if __name__ == '__main__':
    parser = argparse.ArgumentParser()

    parser.add_argument(
        '--host',
        help='contact service host',
        default=DEFAULT_HOST
    )

    parser.add_argument(
        '--port',
        help='contact service port',
        default=DEFAULT_PORT
    )

    args = parser.parse_args()
    params = args.__dict__

    main(params)
