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
        String occurrenceId,
        Paikot paikot,
        /** How much of the amount holds a card's limit until paid (installment principal); null = all of it. */
        BigDecimal principal,
        /** The plan's monthly interest rate, e.g. 1.78805; null = none. */
        BigDecimal monthlyRatePercent
) {
    /** Where the due lands if its account's linked card pays it; null without a link. Dates are yyyy-MM-dd. */
    public record Paikot(String cardId, String cardName, String payFrom, String payBy, String dueDate, BigDecimal fee) {
    }

    /** PUT body linking an account's dues to the card that can pay them; a null cardId removes the link. */
    public record Link(String cardId, BigDecimal feePercent) {
    }
}
