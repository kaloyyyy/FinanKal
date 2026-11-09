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
import java.time.ZoneId;
import java.time.ZoneOffset;
import java.util.UUID;

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

        if (dto.getDate() != null) {
            builder.setOccurredAt(startOfDay(dto));
        }

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
     * Maps an edit of an existing transaction: every entry by id, with its new account, amount and type.
     */
    public FinanceProtos.UpdateTransactionRequest toUpdateProto(
            String userId,
            UUID transactionId,
            CreateTransactionRequestDto dto
    ) {
        FinanceProtos.UpdateTransactionRequest.Builder builder =
                FinanceProtos.UpdateTransactionRequest.newBuilder()
                        .setUserId(userId)
                        .setTransactionId(transactionId.toString())
                        .setDescription(dto.getDescription() == null ? "" : dto.getDescription());

        if (dto.getDate() != null) {
            builder.setOccurredAt(startOfDay(dto));
        }

        for (CreateTransactionEntryDto entry : dto.getEntries()) {
            builder.addEntries(FinanceProtos.TransactionEntry.newBuilder()
                    .setId(entry.getId() == null ? "" : entry.getId())
                    .setAccountId(entry.getAccountId())
                    .setAmount(entry.getAmount().toString())
                    .setType(entry.getType()));
        }

        return builder.build();
    }

    /** Dates from statements and the form are Philippine calendar days. */
    private static Timestamp startOfDay(CreateTransactionRequestDto dto) {
        Instant start = dto.getDate().atStartOfDay(ZoneId.of("Asia/Manila")).toInstant();
        return Timestamp.newBuilder().setSeconds(start.getEpochSecond()).build();
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

        dto.setId(entry.getId());
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