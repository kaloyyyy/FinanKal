package com.finankal.api.dto;

import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Positive;
import lombok.Data;

import java.math.BigDecimal;
import java.util.Objects;

@Data
public class CreateCreditCardRequestDto {

    private String accountId;

    private String accountName;

    @NotNull(message = "Credit limit is required")
    @Positive(message = "Credit limit must be positive")
    private BigDecimal creditLimit;

    @NotNull(message = "Billing day is required")
    private Integer billingDay;

    @NotNull(message = "Payment due days is required")
    private Integer paymentDueDays;

    public boolean hasAccountId() {
        return Objects.nonNull(accountId) && !accountId.trim().isEmpty();
    }

    public boolean hasAccountName() {
        return Objects.nonNull(accountName) && !accountName.trim().isEmpty();
    }

    public boolean isValid() {
        return Objects.nonNull(accountId) || Objects.nonNull(accountName);
    }
}