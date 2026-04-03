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

@Service
public class TransactionService {

    @Autowired
    private FinanceEngineGrpc.FinanceEngineBlockingStub financeEngineStub;

    @Autowired
    private TransactionMapper transactionMapper;

    public TransactionDto createTransaction(CreateTransactionRequestDto requestDto) {
        FinanceProtos.CreateTransactionRequest request = transactionMapper.toProto(requestDto);
        FinanceProtos.CreateTransactionResponse response = financeEngineStub.createTransaction(request);

        // Create DTO with both gRPC response data and request data
        TransactionDto dto = new TransactionDto();
        dto.setId(response.getTransactionId());
        dto.setDescription(requestDto.getDescription());
        dto.setEntries(requestDto.getEntries());

        // Populate additional fields in entries
        for (LedgerEntryDto entry : dto.getEntries()) {
            entry.setTransactionId(dto.getId());
            entry.setDescription(dto.getDescription());
            entry.setCreatedAt(LocalDateTime.now().format(DateTimeFormatter.ISO_LOCAL_DATE_TIME));
        }

        return dto;
    }
}
