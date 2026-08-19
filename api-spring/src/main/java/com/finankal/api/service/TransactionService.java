package com.finankal.api.service;

import com.finankal.api.dto.CreateTransactionRequestDto;
import com.finankal.api.dto.TransactionDto;
import com.finankal.api.finance.FinanceEngineGrpc;
import com.finankal.api.finance.FinanceProtos;
import com.finankal.api.mapper.TransactionMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.UUID;

@Service
public class TransactionService {

    @Autowired
    private FinanceEngineGrpc.FinanceEngineBlockingStub financeEngineStub;

    @Autowired
    private TransactionMapper transactionMapper;


    public TransactionDto createTransaction(
            CreateTransactionRequestDto requestDto
    ) {

        FinanceProtos.CreateTransactionRequest request =
                transactionMapper.toProto(requestDto);

        FinanceProtos.CreateTransactionResponse response =
                financeEngineStub.createTransaction(request);

        return transactionMapper.toDto(response);
    }


    public TransactionDto getTransaction(
            UUID transactionId
    ) {

        FinanceProtos.GetTransactionRequest request =
                FinanceProtos.GetTransactionRequest.newBuilder()
                        .setTransactionId(transactionId.toString())
                        .build();

        FinanceProtos.TransactionResponse response =
                financeEngineStub.getTransaction(request);

        return transactionMapper.toDto(response);
    }


    public List<TransactionDto> getUserTransactions(
            UUID userId,
            int limit,
            int offset
    ) {

        FinanceProtos.GetUserTransactionsRequest request =
                FinanceProtos.GetUserTransactionsRequest.newBuilder()
                        .setUserId(userId.toString())
                        .setLimit(limit)
                        .setOffset(offset)
                        .build();

        FinanceProtos.GetUserTransactionsResponse response =
                financeEngineStub.getUserTransactions(request);

        return response.getTransactionsList()
                .stream()
                .map(transactionMapper::toDto)
                .toList();
    }
}