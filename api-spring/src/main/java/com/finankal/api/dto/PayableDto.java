package com.finankal.api.dto;

import java.math.BigDecimal;

/** One unpaid installment of a loan or card. dueDate and scheduledDate are yyyy-MM-dd; numbers are null when unknown. */
public record PayableDto(
        String name,
        String accountId,
        String accountName,
        String dueDate,
        String scheduledDate,
        BigDecimal amount,
        Integer installmentNumber,
        Integer installments,
        String sourceAccountId,
        String scheduleId,
        String occurrenceId
) {
}
