import grpc

import contact_pb2
import contact_pb2_grpc


def main():
    with grpc.insecure_channel('localhost:50051') as channel:
        stub = contact_pb2_grpc.ContactStub(channel)

        message = contact_pb2.PostMessageRequest()
        message.message = 'This does not please me.'

        print(stub.PostMessage(message))


if __name__ == '__main__':
    main()
