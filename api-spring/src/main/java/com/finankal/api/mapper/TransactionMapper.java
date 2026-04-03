package com.finankal.api.mapper;

import com.finankal.api.dto.CreateTransactionRequestDto;
import com.finankal.api.dto.LedgerEntryDto;
import com.finankal.api.dto.TransactionDto;
import com.finankal.api.finance.FinanceProtos;
import org.springframework.stereotype.Component;

import java.math.BigDecimal;
import java.util.List;
import java.util.stream.Collectors;

@Component
public class TransactionMapper {

    public FinanceProtos.CreateTransactionRequest toProto(CreateTransactionRequestDto dto) {
        FinanceProtos.CreateTransactionRequest.Builder builder = FinanceProtos.CreateTransactionRequest.newBuilder()
                .setDescription(dto.getDescription());

        for (LedgerEntryDto entry : dto.getEntries()) {
            FinanceProtos.Entry protoEntry = FinanceProtos.Entry.newBuilder()
                    .setAccountId(entry.getAccountId())
                    .setAmount(entry.getAmount().toString())
                    .setType(entry.getType())
                    .build();
            builder.addEntries(protoEntry);
        }

        return builder.build();
    }

    public TransactionDto toDto(FinanceProtos.CreateTransactionResponse response) {
        TransactionDto dto = new TransactionDto();
        dto.setId(response.getTransactionId());
        // Description and entries not in response, so set to null
        return dto;
    }
}
