package com.finankal.api.dto;

import lombok.Data;
import java.math.BigDecimal;
import java.time.Instant;

@Data
public class RecordCreditCardTransactionRequestDto {
    private String cardId;
    private BigDecimal amount;
    private String description;
    private Instant purchaseDate;
}

