package com.finankal.api.dto;

import java.math.BigDecimal;

/** Money that moved from one account to another, summed over `transactions` transactions. */
public record MoneyFlowDto(
        String fromAccountId,
        String toAccountId,
        BigDecimal amount,
        int transactions
) {
}
