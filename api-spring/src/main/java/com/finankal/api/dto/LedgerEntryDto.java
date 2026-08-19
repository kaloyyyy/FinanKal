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
}