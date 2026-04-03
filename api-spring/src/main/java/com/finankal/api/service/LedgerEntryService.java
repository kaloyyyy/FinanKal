package com.finankal.api.service;

import com.finankal.api.dto.LedgerEntryDto;
import com.finankal.api.finance.FinanceEngineGrpc;
import com.finankal.api.finance.FinanceProtos;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.math.BigDecimal;
import java.util.List;
import java.util.stream.Collectors;

@Service
public class LedgerEntryService {

    @Autowired
    private FinanceEngineGrpc.FinanceEngineBlockingStub financeEngineStub;

    public List<LedgerEntryDto> getLedgerEntries(String accountId) {
        FinanceProtos.GetLedgerEntriesRequest request = FinanceProtos.GetLedgerEntriesRequest.newBuilder()
                .setAccountId(accountId)
                .build();

        FinanceProtos.GetLedgerEntriesResponse response = financeEngineStub.getLedgerEntries(request);

        return response.getEntriesList().stream()
                .map(this::mapToDto)
                .collect(Collectors.toList());
    }

    private LedgerEntryDto mapToDto(FinanceProtos.LedgerEntry protoEntry) {
        LedgerEntryDto dto = new LedgerEntryDto();
        dto.setAccountId(protoEntry.getAccountId());
        dto.setAmount(new BigDecimal(protoEntry.getAmount()));
        dto.setType(protoEntry.getType());
        dto.setDescription(protoEntry.getDescription());
        dto.setTransactionId(protoEntry.getTransactionId());
        dto.setCreatedAt(protoEntry.getCreatedAt());
        return dto;
    }
}
