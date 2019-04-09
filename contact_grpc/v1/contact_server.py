import grpc
import time
from concurrent import futures

import contact_pb2
import contact_pb2_grpc


class ContactServer(contact_pb2_grpc.ContactServicer):

    def PostMessage(self, request, context):
        resp = contact_pb2.PostMessageResponse()
        resp.request.CopyFrom(request)
        return resp

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    contact_pb2_grpc.add_ContactServicer_to_server(ContactServer(), server)

    server.add_insecure_port('[::]:50051')
    server.start()
    try:
        while True:
            time.sleep(1000000)
    except KeyboardInterrupt:
        server.stop(0)

if __name__ == '__main__':
    serve()
