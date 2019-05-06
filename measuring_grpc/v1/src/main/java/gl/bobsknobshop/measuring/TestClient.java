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

import com.google.protobuf.Any;
import gl.bobsknobshop.contact.v1.PostMessageRequest;
import gl.bobsknobshop.measuring.v1.MeasuringGrpc;
import gl.bobsknobshop.measuring.v1.PostServiceCallRequest;
import gl.bobsknobshop.measuring.v1.PostServiceCallResponse;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.StatusRuntimeException;

import java.util.concurrent.TimeUnit;
import java.util.logging.Level;
import java.util.logging.Logger;


public class TestClient {
  private static final Logger LOG = Logger.getLogger(
      TestClient.class.getName());

  public static void main(String[] args) throws Exception {
    final int port;
    if (args.length > 1) {
      port = Integer.valueOf(args[0]);
    } else {
      port = MeasuringServer.DEFAULT_PORT;
    }

    final String host;
    if (args.length > 2) {
      host = args[1];
    } else {
      host = "localhost";
    }

    try (MeasuringClient client = MeasuringClient.of(host, port)) {

      final PostServiceCallRequest request = createRequest();
      final PostServiceCallResponse response = client.postServiceCall(request);

      LOG.info(response.toString());
    } catch (StatusRuntimeException e) {
      LOG.log(Level.WARNING, "RPC failed: {0}", e.getStatus());
    }
  }

  private static PostServiceCallRequest createRequest() {
    final PostServiceCallRequest request = PostServiceCallRequest.newBuilder()
        .setServiceName("messaging")
        .setRpcName("postmessage")
        .setRequest(Any.newBuilder().build())
        .setResponse(Any.newBuilder().build())
        .build();

    return request;
  }


  static class MeasuringClient implements AutoCloseable {

    private final ManagedChannel channel;
    private final MeasuringGrpc.MeasuringBlockingStub
        blockingStub;

    MeasuringClient(ManagedChannel channel) {
      this.channel = channel;
      blockingStub = MeasuringGrpc.newBlockingStub(channel);
    }

    public static MeasuringClient of(String host, int port) {
      final ManagedChannel managedChannel = ManagedChannelBuilder
          .forAddress(host, port)
          .usePlaintext()
          .build();
      return new MeasuringClient(managedChannel);
    }

    public void close() throws InterruptedException {
      channel.shutdown().awaitTermination(5, TimeUnit.SECONDS);
    }

    public PostServiceCallResponse postServiceCall(PostServiceCallRequest request) {
      return blockingStub.postServiceCall(request);
    }
  }
}
