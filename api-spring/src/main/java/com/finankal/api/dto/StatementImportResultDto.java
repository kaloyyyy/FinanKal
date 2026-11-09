package com.finankal.api.dto;

import java.util.List;

/**
 * The outcome for one uploaded statement file.
 *
 * bank / accountName: what the statement was detected as and where it went (null when it failed before that).
 * imported: transactions recorded from the statement (the whole statement is imported or nothing is).
 * skipped: statement rows that looked like transactions but whose direction couldn't be determined.
 * duplicates: rows left out because the account already had them (an overlapping statement or a hand-entered transaction).
 * transfers: rows joined with the other end of a transfer between the user's own accounts.
 * warnings: where FinanKal's balance and the account's statements disagree (missing month, misread row, manual entry).
 * error: why the file was not imported (ALREADY_IMPORTED or FAILED).
 */
public record StatementImportResultDto(
        String fileName,
        Status status,
        String bank,
        String accountName,
        int imported,
        List<String> skipped,
        List<String> duplicates,
        List<String> transfers,
        List<String> warnings,
        String error
) {
    public enum Status { IMPORTED, ALREADY_IMPORTED, FAILED }

    public static StatementImportResultDto notImported(String fileName, Status status, String bank, String error) {
        return new StatementImportResultDto(fileName, status, bank, null, 0, List.of(), List.of(), List.of(), List.of(), error);
    }
}
