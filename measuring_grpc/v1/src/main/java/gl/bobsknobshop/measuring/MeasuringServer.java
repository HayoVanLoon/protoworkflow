/*
 * Copyright 2019 Hayo van Loon
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Code adapted from the grpc-java example available at https://github.com/grpc/grpc-java/

package gl.bobsknobshop.measuring;

import gl.bobsknobshop.measuring.v1.MeasuringGrpc;
import gl.bobsknobshop.measuring.v1.PostServiceCallRequest;
import gl.bobsknobshop.measuring.v1.PostServiceCallResponse;
import io.grpc.Server;
import io.grpc.ServerBuilder;
import io.grpc.stub.StreamObserver;

import java.io.IOException;
import java.util.logging.Logger;


public class MeasuringServer {

  private static final Logger LOG = Logger.getLogger(
      MeasuringServer.class.getName());

  /** Default listening port */
  static final int DEFAULT_PORT = 8080;

  private final int port;
  private Server server;

  private MeasuringServer(int port) {
    this.port = port;
  }

  static MeasuringServer defaultInstance() {
    return new MeasuringServer(DEFAULT_PORT);
  }

  static MeasuringServer on(int port) {
    return new MeasuringServer(port);
  }

  /**
   * Main launches the server from the command line.
   */
  public static void main(String[] args)
      throws IOException, InterruptedException {

    final MeasuringServer server;
    if (args.length > 1) {
      server = MeasuringServer.on(Integer.valueOf(args[0]));
    } else {
      server = MeasuringServer.defaultInstance();
    }

    server.start();
    server.blockUntilShutdown();
  }

  private void start() throws IOException {
    server = ServerBuilder.forPort(port)
        .addService(new MeasuringImpl())
        .build()
        .start();
    LOG.info("Server started, listening on " + port);
    Runtime.getRuntime().addShutdownHook(new Thread(() -> {
      // Use stderr here since the LOG may have been reset by its JVM shutdown hook.
      System.err
          .println("*** shutting down gRPC server since JVM is shutting down");
      MeasuringServer.this.stop();
      System.err.println("*** server shut down");
    }));
  }

  private void stop() {
    if (server != null) {
      server.shutdown();
    }
  }

  /**
   * Await termination on the main thread since the grpc library uses daemon
   * threads.
   */
  private void blockUntilShutdown() throws InterruptedException {
    if (server != null) {
      server.awaitTermination();
    }
  }


  static class MeasuringImpl extends MeasuringGrpc.MeasuringImplBase {

    @Override
    public void postServiceCall(PostServiceCallRequest request,
                                StreamObserver<PostServiceCallResponse> responseObserver) {

      LOG.info("Received a PostServiceCall request " + request);

      final PostServiceCallResponse.Builder resp =
          PostServiceCallResponse.newBuilder();

      responseObserver.onNext(resp.build());
      responseObserver.onCompleted();
    }
  }
}
