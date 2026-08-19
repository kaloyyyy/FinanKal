package com.finankal.api.mapper;

import com.finankal.api.dto.CreateTransactionEntryDto;
import com.finankal.api.dto.CreateTransactionRequestDto;
import com.finankal.api.dto.LedgerEntryDto;
import com.finankal.api.dto.TransactionDto;
import com.finankal.api.finance.FinanceProtos;
import com.google.protobuf.Timestamp;
import org.springframework.stereotype.Component;

import java.math.BigDecimal;
import java.time.Instant;
import java.time.OffsetDateTime;
import java.time.ZoneOffset;

@Component
public class TransactionMapper {

    /**
     * Maps API create request to protobuf request.
     */
    public FinanceProtos.CreateTransactionRequest toProto(
            CreateTransactionRequestDto dto
    ) {
        FinanceProtos.CreateTransactionRequest.Builder builder =
                FinanceProtos.CreateTransactionRequest.newBuilder()
                        .setDescription(dto.getDescription());

        for (CreateTransactionEntryDto entry : dto.getEntries()) {

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
     *
     * Currently CreateTransactionResponse only contains
     * the transaction ID.
     */
    public TransactionDto toDto(
            FinanceProtos.CreateTransactionResponse response
    ) {
        TransactionDto dto = new TransactionDto();

        dto.setId(response.getTransactionId());

        return dto;
    }


    /**
     * Maps a full protobuf transaction to API DTO.
     *
     * Used by:
     * - GetTransaction
     * - GetUserTransactions
     */
    public TransactionDto toDto(
            FinanceProtos.TransactionResponse response
    ) {
        TransactionDto dto = new TransactionDto();

        dto.setId(response.getId());
        dto.setDescription(response.getDescription());

        if (response.hasCreatedAt()) {
            dto.setCreatedAt(toOffsetDateTime(response.getCreatedAt()));
        }

        dto.setEntries(
                response.getEntriesList()
                        .stream()
                        .map(this::toLedgerEntryDto)
                        .toList()
        );

        return dto;
    }


    /**
     * Maps protobuf transaction entry to API DTO.
     */
    private LedgerEntryDto toLedgerEntryDto(
            FinanceProtos.TransactionEntry entry
    ) {
        LedgerEntryDto dto = new LedgerEntryDto();

        dto.setAccountId(entry.getAccountId());
        dto.setAccountName(entry.getAccountName());
        dto.setAccountType(entry.getAccountType());
        dto.setAmount(new BigDecimal(entry.getAmount()));
        dto.setType(entry.getType());
        dto.setTransactionId(entry.getTransactionId());

        return dto;
    }


    /**
     * Converts protobuf Timestamp to OffsetDateTime.
     */
    private OffsetDateTime toOffsetDateTime(
            Timestamp timestamp
    ) {
        return Instant.ofEpochSecond(
                timestamp.getSeconds(),
                timestamp.getNanos()
        ).atOffset(ZoneOffset.UTC);
    }
}