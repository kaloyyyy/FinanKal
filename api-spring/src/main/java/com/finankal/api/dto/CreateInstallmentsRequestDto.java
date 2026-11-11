package com.finankal.api.dto;

import java.math.BigDecimal;
import java.util.List;

/** A loan's repayment schedule: installments in due-date order (yyyy-MM-dd), paid from sourceAccountId into accountId. */
public record CreateInstallmentsRequestDto(
        String name,
        String sourceAccountId,
        String accountId,
        List<Installment> installments,
        int firstInstallmentNumber
) {
    public record Installment(String dueDate, BigDecimal amount) {
    }
}
