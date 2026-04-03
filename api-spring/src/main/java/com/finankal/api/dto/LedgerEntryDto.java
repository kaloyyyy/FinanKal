package com.finankal.api.dto;

import lombok.Data;
import java.math.BigDecimal;

@Data
public class LedgerEntryDto {
    private String accountId;
    private BigDecimal amount;
    private String type; // DEBIT or CREDIT
    private String description;
    private String transactionId;
    private String createdAt;
}
