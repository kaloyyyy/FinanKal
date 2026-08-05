package com.finankal.api.dto;

import lombok.Data;
import java.math.BigDecimal;

@Data
public class PayCreditCardStatementRequestDto {
    private String statementId;
    private String cardId;
    private String paymentAccountId;
    private BigDecimal amount;
    private String description;
}

