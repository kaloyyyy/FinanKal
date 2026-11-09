package com.finankal.api.dto;

import lombok.Data;

import java.math.BigDecimal;

@Data
public class CreateTransactionEntryDto {

    /** The entry being changed; only used when updating a transaction. */
    private String id;
    private String accountId;
    private BigDecimal amount;
    private String type;
}