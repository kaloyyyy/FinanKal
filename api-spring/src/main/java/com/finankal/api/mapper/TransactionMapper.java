package com.finankal.api.mapper;

import com.finankal.api.dto.CreateTransactionRequestDto;
import com.finankal.api.dto.LedgerEntryDto;
import com.finankal.api.dto.TransactionDto;
import com.finankal.api.finance.FinanceProtos;
import org.springframework.stereotype.Component;

import java.math.BigDecimal;

@Component
public class TransactionMapper {

    public FinanceProtos.CreateTransactionRequest toProto(
            CreateTransactionRequestDto dto
    ) {
        FinanceProtos.CreateTransactionRequest.Builder builder =
                FinanceProtos.CreateTransactionRequest.newBuilder()
                        .setDescription(dto.getDescription());

        for (LedgerEntryDto entry : dto.getEntries()) {

            FinanceProtos.Entry protoEntry =
                    FinanceProtos.Entry.newBuilder()
                            .setAccountId(entry.getAccountId())
                            .setAmount(entry.getAmount().toString())
                            .setType(entry.getType())
                            .build();

            builder.addEntries(protoEntry);
        }

        return builder.build();
    }


    /**
     * Response after creating a transaction.
     */
    public TransactionDto toDto(
            FinanceProtos.CreateTransactionResponse response
    ) {
        TransactionDto dto = new TransactionDto();

        dto.setId(response.getTransactionId());

        return dto;
    }


    /**
     * Full transaction returned by GetTransaction
     * and GetUserTransactions.
     */
    public TransactionDto toDto(
            FinanceProtos.TransactionResponse response
    ) {
        TransactionDto dto = new TransactionDto();

        dto.setId(response.getId());
        dto.setDescription(response.getDescription());

        dto.setEntries(
                response.getEntriesList()
                        .stream()
                        .map(this::toLedgerEntryDto)
                        .toList()
        );

        return dto;
    }


    /**
     * Maps protobuf transaction entry to the API DTO.
     */
    private LedgerEntryDto toLedgerEntryDto(
            FinanceProtos.TransactionEntry entry
    ) {
        LedgerEntryDto dto = new LedgerEntryDto();

        dto.setAccountId(entry.getAccountId());
        dto.setAmount(new BigDecimal(entry.getAmount()));
        dto.setType(entry.getType());
        dto.setTransactionId(entry.getTransactionId());

        return dto;
    }
}