package com.finankal.api.dto;

import java.math.BigDecimal;

/** A card's editable terms; dueRule and statementShift take the engine's values (e.g. DAYS, NEXT_BUSINESS_DAY). */
public record CreditCardTermsDto(BigDecimal creditLimit, int billingDay, int paymentDueDays,
                                 String dueRule, String statementShift) {}
