package com.finankal.api.service;

import com.finankal.api.dto.CreateTransactionRequestDto;
import com.finankal.api.dto.TransactionDto;
import com.finankal.api.finance.FinanceEngineGrpc;
import com.finankal.api.finance.FinanceProtos;
import com.finankal.api.mapper.TransactionMapper;
import io.grpc.StatusRuntimeException;
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


    /** Edits one of the user's transactions; a rejected edit (doesn't balance, not theirs) is a 400 with the reason. */
    public void updateTransaction(String userId, UUID transactionId, CreateTransactionRequestDto requestDto) {
        try {
            financeEngineStub.updateTransaction(transactionMapper.toUpdateProto(userId, transactionId, requestDto));
        } catch (StatusRuntimeException e) {
            throw rejected(e);
        }
    }

    public void deleteTransaction(String userId, UUID transactionId) {
        try {
            financeEngineStub.deleteTransaction(FinanceProtos.DeleteTransactionRequest.newBuilder()
                    .setUserId(userId)
                    .setTransactionId(transactionId.toString())
                    .build());
        } catch (StatusRuntimeException e) {
            throw rejected(e);
        }
    }

    private static RuntimeException rejected(StatusRuntimeException e) {
        return switch (e.getStatus().getCode()) {
            case INVALID_ARGUMENT -> new IllegalArgumentException(e.getStatus().getDescription());
            case NOT_FOUND -> new IllegalArgumentException("Transaction not found");
            default -> e;
        };
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