package com.finankal.api.dto;

import lombok.Data;

import java.time.OffsetDateTime;
import java.util.List;

@Data
public class TransactionDto {

    private String id;
    private String description;
    private OffsetDateTime createdAt;

    private List<LedgerEntryDto> entries;
}