package com.finankal.api.dto;

import lombok.Data;

import java.math.BigDecimal;

@Data
public class LedgerEntryDto {

    private String id;

    private String accountId;
    private String accountName;
    private String accountType;

    private BigDecimal amount;
    private String type;

    private String transactionId;

    private String description;

    /** RFC 3339, as sent by the engine. */
    private String createdAt;
}