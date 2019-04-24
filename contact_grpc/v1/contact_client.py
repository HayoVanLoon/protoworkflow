import argparse
import grpc
from bobsknobshop.contact.v1 import contact_pb2
from bobsknobshop.contact.v1 import contact_pb2_grpc
from google.protobuf import json_format


def main(params):
    with grpc.insecure_channel(params['host'] + ':' + params['port']) as channel:
        stub = contact_pb2_grpc.ContactStub(channel)

        message = contact_pb2.PostMessageRequest()
        message.message = 'This does not please me.'

        resp = stub.PostMessage(message)
        print(json_format.MessageToJson(resp))


if __name__ == '__main__':
    parser = argparse.ArgumentParser()

    parser.add_argument(
        '--host',
        help='contact service host',
        default='localhost'
    )

    parser.add_argument(
        '--port',
        help='contact service port',
        default=8083
    )

    args = parser.parse_args()
    params = args.__dict__

    main(params)
