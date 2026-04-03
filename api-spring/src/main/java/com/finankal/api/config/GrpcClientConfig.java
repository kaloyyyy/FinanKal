package com.finankal.api.config;

import com.finankal.api.finance.FinanceEngineGrpc;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class GrpcClientConfig {

    @Bean
    public ManagedChannel managedChannel() {
        return ManagedChannelBuilder.forAddress("localhost", 50051) // Assuming the Go engine runs on 50051
                .usePlaintext()
                .build();
    }

    @Bean
    public FinanceEngineGrpc.FinanceEngineBlockingStub financeEngineStub(ManagedChannel channel) {
        return FinanceEngineGrpc.newBlockingStub(channel);
    }
}
