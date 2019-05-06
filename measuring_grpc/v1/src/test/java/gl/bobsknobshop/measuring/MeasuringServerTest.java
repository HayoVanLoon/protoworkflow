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

import gl.bobsknobshop.measuring.MeasuringServer.MeasuringImpl;
import gl.bobsknobshop.measuring.v1.MeasuringGrpc;
import gl.bobsknobshop.measuring.v1.PostServiceCallRequest;
import gl.bobsknobshop.measuring.v1.PostServiceCallResponse;
import io.grpc.inprocess.InProcessChannelBuilder;
import io.grpc.inprocess.InProcessServerBuilder;
import io.grpc.testing.GrpcCleanupRule;
import org.junit.Assert;
import org.junit.Before;
import org.junit.Rule;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.junit.runners.JUnit4;


@RunWith(JUnit4.class)
public class MeasuringServerTest {

  private MeasuringGrpc.MeasuringBlockingStub blockingStub;

  @Rule
  public final GrpcCleanupRule grpcCleanup = new GrpcCleanupRule();

  @Before
  public void setUp() throws Exception {
    String serverName = InProcessServerBuilder.generateName();
    // Create a server, add service, start, and register for automatic graceful shutdown.
    grpcCleanup.register(InProcessServerBuilder
        .forName(serverName).directExecutor().addService(new MeasuringImpl())
        .build().start());

    blockingStub = MeasuringGrpc.newBlockingStub(
        // Create a client channel and register for automatic graceful shutdown.
        grpcCleanup.register(
            InProcessChannelBuilder.forName(serverName).directExecutor()
                .build()));

  }

  @Test
  public void happy() {
    PostServiceCallRequest request = PostServiceCallRequest.newBuilder()
        .build();

    PostServiceCallResponse response = blockingStub.postServiceCall(request);

    Assert.assertEquals(PostServiceCallResponse.newBuilder().build(), response);
  }
}
