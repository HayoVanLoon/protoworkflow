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
