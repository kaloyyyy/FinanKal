package com.finankal.api.dto;

/** An imported statement. bank is the code (MARIBANK, UNIONBANK, ...); dates are yyyy-MM-dd; importedAt is RFC 3339. */
public record BankStatementDto(
        String id,
        String accountId,
        String bank,
        String periodStart,
        String periodEnd,
        String importedAt
) {
}
