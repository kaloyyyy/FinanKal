package com.finankal.api.dto;

import java.math.BigDecimal;

/** One unpaid installment of a loan or card. dueDate is yyyy-MM-dd; installment numbers are null when unknown. */
public record PayableDto(
        String name,
        String accountId,
        String accountName,
        String dueDate,
        BigDecimal amount,
        Integer installmentNumber,
        Integer installments
) {
}
