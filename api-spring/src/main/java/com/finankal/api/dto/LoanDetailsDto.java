package com.finankal.api.dto;

import java.math.BigDecimal;

public record LoanDetailsDto(String accountId, String name, String bookedOn, BigDecimal dailyRate,
                             String principalOn, BigDecimal principal) {}
