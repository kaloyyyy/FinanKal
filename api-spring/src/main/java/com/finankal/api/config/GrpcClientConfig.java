package com.finankal.api.config;

import com.finankal.api.finance.FinanceEngineGrpc;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class GrpcClientConfig {

    @Value("${grpc.client.host}")
    private String grpcHost;

    @Value("${grpc.client.port}")
    private int grpcPort;

    @Bean
    public ManagedChannel managedChannel() {
        return ManagedChannelBuilder
                .forAddress(grpcHost, grpcPort)
                .usePlaintext()
                .build();
    }

    @Bean
    public FinanceEngineGrpc.FinanceEngineBlockingStub financeEngineStub(
            ManagedChannel channel
    ) {
        return FinanceEngineGrpc.newBlockingStub(channel);
    }
}