package com.finankal.api.dto;

import java.math.BigDecimal;
import java.time.OffsetDateTime;
import java.util.UUID;

public record RecentTransactionDto(
        UUID id,
        String description,
        BigDecimal amount,
        String direction,
        String accountName,
        OffsetDateTime createdAt
) {
}