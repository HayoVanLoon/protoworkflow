import argparse
import logging

import grpc
import time
from bobsknobshop.contact.v1 import contact_pb2
from bobsknobshop.contact.v1 import contact_pb2_grpc
from bobsknobshop.messaging.v1 import messaging_pb2
from bobsknobshop.messaging.v1 import messaging_pb2_grpc
from concurrent import futures

DEFAULT_PORT = 8080
DEFAULT_TIMEOUT = 10


logging.basicConfig(level=logging.DEBUG)
LOGGER = logging.getLogger(__name__)
LOGGER.setLevel(logging.INFO)


class ContactServer(contact_pb2_grpc.ContactServicer):

    def __init__(self, messaging_host, messaging_port):
        super(ContactServer, self).__init__()
        self.messaging_target = messaging_host + ':' + str(messaging_port)

    def PostMessage(self, request, context):
        message_req = messaging_pb2.PostMessageRequest()
        message_req.customer_message.body = request.message
        message_req.customer_message.sender.name = 'foo_name'
        message_req.customer_message.sender.email = 'foo@example.com'
        message_req.customer_message.sender.name = 'Foo Bar'
        message_req.customer_message.sender.name = 'Mrs.'

        resp = contact_pb2.PostMessageResponse()

        with grpc.insecure_channel(self.messaging_target) as channel:
            stub = messaging_pb2_grpc.MessagingStub(channel)

            try:
                message_resp = stub.PostMessage(message_req,
                                                timeout=DEFAULT_TIMEOUT)
            except Exception as ex:
                LOGGER.warning(ex)

        return resp


def serve():
    parser = argparse.ArgumentParser()

    parser.add_argument(
        '--port',
        help='server will listen on this port',
        default=DEFAULT_PORT
    )
    parser.add_argument(
        '--messaging_host',
        help='messaging service host',
        default='messaging-service'
    )
    parser.add_argument(
        '--messaging_port',
        help='messaging service port',
        default=DEFAULT_PORT
    )

    args = parser.parse_args()
    params = args.__dict__

    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    contact_server = ContactServer(params['messaging_host'],
                                   params['messaging_port'])
    contact_pb2_grpc.add_ContactServicer_to_server(contact_server, server)

    server.add_insecure_port('[::]:%s' % params['port'])
    server.start()

    LOGGER.info('server started with messaging target {}'
                .format(contact_server.messaging_target))

    try:
        while True:
            time.sleep(1000000)
    except KeyboardInterrupt:
        server.stop(0)


if __name__ == '__main__':
    serve()
