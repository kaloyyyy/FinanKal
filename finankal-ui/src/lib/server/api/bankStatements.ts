const API_URL = "http://localhost:8080";

export const BANKS = [
    { value: "GCASH", label: "GCash" },
    { value: "MAYA", label: "Maya" },
    { value: "BDO", label: "BDO" },
    { value: "BPI", label: "BPI" },
    { value: "GOTYME", label: "GoTyme" },
    { value: "MARIBANK", label: "MariBank" },
    { value: "UNIONBANK", label: "UnionBank" },
    { value: "UNIONBANK_CARD", label: "UnionBank Credit Card" }
] as const;

export type Bank = (typeof BANKS)[number]["value"];

/** The outcome for one uploaded file. */
export interface StatementImportResult {
    fileName: string;
    status: "IMPORTED" | "ALREADY_IMPORTED" | "FAILED";
    /** Detected or chosen bank; null when the file failed before that. */
    bank: string | null;
    /** The account the statement went into (found by its account number, or created). */
    accountName: string | null;
    imported: number;
    skipped: string[];
    /** Rows left out because the account already had them. */
    duplicates: string[];
    /** Rows joined with the other end of a transfer between your own accounts. */
    transfers: string[];
    /** Where FinanKal's balance and the account's statements disagree. */
    warnings: string[];
    /** Why the file wasn't imported. */
    error: string | null;
}

/** An imported statement; dates are yyyy-MM-dd. */
export interface BankStatement {
    id: string;
    accountId: string;
    bank: string;
    periodStart: string;
    periodEnd: string;
    importedAt: string;
}

/** GET /api/bank-statements?userId=... (oldest period first) */
export async function getBankStatements(userId: string): Promise<BankStatement[]> {
    const response = await fetch(`${API_URL}/api/bank-statements?userId=${encodeURIComponent(userId)}`);
    if (!response.ok) {
        throw new Error(`Couldn't load statements (${response.status})`);
    }
    return response.json();
}

/**
 * POST /api/bank-statements/import (multipart/form-data)
 * Without bank / sourceAccountId, both are detected from each statement.
 */
export async function importBankStatements(payload: {
    files: File[];
    userId: string;
    bank?: Bank;
    sourceAccountId?: string;
    password?: string;
}): Promise<StatementImportResult[]> {
    const body = new FormData();
    for (const file of payload.files) {
        body.append("files", file);
    }
    body.set("userId", payload.userId);
    if (payload.bank) {
        body.set("bank", payload.bank);
    }
    if (payload.sourceAccountId) {
        body.set("sourceAccountId", payload.sourceAccountId);
    }
    if (payload.password) {
        body.set("password", payload.password);
    }

    const response = await fetch(`${API_URL}/api/bank-statements/import`, {
        method: "POST",
        body
    });

    if (!response.ok) {
        // Spring's GlobalExceptionHandler returns { message }; 400s carry a user-readable reason.
        const error = await response.json().catch(() => null);
        throw new StatementImportError(response.status, error?.message ?? "The statements couldn't be imported.");
    }

    return response.json();
}

export class StatementImportError extends Error {
    constructor(public status: number, message: string) {
        super(message);
    }
}
