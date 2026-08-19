package com.finankal.api.service;

import com.finankal.api.dto.CreateTransactionRequestDto;
import com.finankal.api.dto.LedgerEntryDto;
import com.finankal.api.dto.TransactionDto;
import com.finankal.api.finance.FinanceEngineGrpc;
import com.finankal.api.finance.FinanceProtos;
import com.finankal.api.mapper.TransactionMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.List;
import java.util.UUID;

@Service
public class TransactionService {

    @Autowired
    private FinanceEngineGrpc.FinanceEngineBlockingStub financeEngineStub;

    @Autowired
    private TransactionMapper transactionMapper;


    /**
     * Create transaction
     */
    public TransactionDto createTransaction(
            CreateTransactionRequestDto requestDto
    ) {

        FinanceProtos.CreateTransactionRequest request =
                transactionMapper.toProto(requestDto);

        FinanceProtos.CreateTransactionResponse response =
                financeEngineStub.createTransaction(request);

        TransactionDto dto = new TransactionDto();

        dto.setId(response.getTransactionId());
        dto.setDescription(requestDto.getDescription());
        dto.setEntries(requestDto.getEntries());

        for (LedgerEntryDto entry : dto.getEntries()) {

            entry.setTransactionId(dto.getId());
            entry.setDescription(dto.getDescription());
            entry.setCreatedAt(
                    LocalDateTime.now()
                            .format(DateTimeFormatter.ISO_LOCAL_DATE_TIME)
            );
        }

        return dto;
    }


    /**
     * Get a specific transaction
     */
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



}