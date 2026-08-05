package com.finankal.api.dto;

import lombok.Data;
import java.math.BigDecimal;
import java.time.LocalDate;

@Data
public class CreditCardStatementDto {
    private String statementId;
    private String creditCardId;
    private LocalDate startDate;
    private LocalDate endDate;
    private LocalDate statementDate;
    private LocalDate dueDate;
    private BigDecimal totalAmount;
    private String status;
}

