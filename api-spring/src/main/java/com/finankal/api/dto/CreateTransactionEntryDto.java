package com.finankal.api.dto;

import lombok.Data;

import java.math.BigDecimal;

@Data
public class CreateTransactionEntryDto {

    private String accountId;
    private BigDecimal amount;
    private String type;
}