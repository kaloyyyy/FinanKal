const API_URL = "http://localhost:8080";

/**
 * Create a new double-entry transaction.
 */
export async function createTransaction(payload: {
    description: string;
    /** yyyy-MM-dd; omitted = now */
    date?: string;
    entries: {
        accountId: string;
        amount: number;
        type: "DEBIT" | "CREDIT";
    }[];
}) {
    const response = await fetch(
        `${API_URL}/api/transactions`,
        {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify(payload)
        }
    );

    if (!response.ok) {
        const errorText = await response.text();

        throw new Error(
            `Spring API returned ${response.status}: ${errorText}`
        );
    }

    return response.json();
}


/**
 * Recent transaction returned by:
 *
 * GET /api/users/{userId}/transactions/recent
 */
export type RecentTransaction = {
    id: string;
    description: string;
    amount: number;
    direction: "IN" | "OUT" | "TRANSFER";
    accountName: string;
    createdAt: string;
};


/**
 * Get the most recent transactions for a user.
 */
export async function getRecentTransactions(
    userId: string,
    limit = 5
): Promise<RecentTransaction[]> {

    const safeLimit = Math.min(Math.max(limit, 1), 20);

    const response = await fetch(
        `${API_URL}/api/users/${userId}/transactions/recent?limit=${safeLimit}`
    );

    if (!response.ok) {
        const errorText = await response.text();

        throw new Error(
            `Failed to fetch recent transactions: ${response.status}: ${errorText}`
        );
    }

    return response.json();
}


/**
 * Ledger entry belonging to a transaction.
 */
export type TransactionEntry = {
    id?: string;
    accountId: string;
    accountName?: string;
    amount: number;
    type: "DEBIT" | "CREDIT";
    description?: string | null;
    transactionId?: string | null;
    createdAt?: string | null;
};


/**
 * Transaction returned by:
 *
 * GET /api/users/{userId}/transactions
 */
export interface Transaction {
    id: string;
    description: string;
    createdAt: string;
    entries: TransactionEntry[];
}


export interface TransactionsResponse {
    transactions: Transaction[];
    total: number;
    limit: number;
    offset: number;
}
/**
 * Get all transactions for a user.
 */
export async function getUserTransactions(
    userId: string
): Promise<TransactionsResponse> {

    const response = await fetch(
        `${API_URL}/api/users/${userId}/transactions`
    );

    if (!response.ok) {
        const errorText = await response.text();

        throw new Error(
            `Failed to fetch transactions: ${response.status}: ${errorText}`
        );
    }

    return response.json();
}


/**
 * Delete all of a user's transactions. Accounts are kept.
 *
 * DELETE /api/users/{userId}/transactions
 */
export async function clearUserTransactions(
    userId: string
): Promise<{ deletedTransactions: number }> {

    const response = await fetch(
        `${API_URL}/api/users/${userId}/transactions`,
        { method: "DELETE" }
    );

    if (!response.ok) {
        const errorText = await response.text();

        throw new Error(
            `Failed to clear transactions: ${response.status}: ${errorText}`
        );
    }

    return response.json();
}


/**
 * Edit a transaction: its description, day and every entry (by id) with its account and amount.
 *
 * PUT /api/transactions/{id}?userId=...
 */
export async function updateTransaction(
    userId: string,
    transactionId: string,
    payload: {
        description: string;
        /** yyyy-MM-dd */
        date: string;
        entries: { id: string; accountId: string; amount: number; type: "DEBIT" | "CREDIT" }[];
    }
): Promise<void> {
    const response = await fetch(
        `${API_URL}/api/transactions/${transactionId}?userId=${encodeURIComponent(userId)}`,
        {
            method: "PUT",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(payload)
        }
    );
    if (!response.ok) {
        throw new Error(await rejection(response, "Couldn't save the transaction."));
    }
}


/**
 * Delete a transaction.
 *
 * DELETE /api/transactions/{id}?userId=...
 */
export async function deleteTransaction(userId: string, transactionId: string): Promise<void> {
    const response = await fetch(
        `${API_URL}/api/transactions/${transactionId}?userId=${encodeURIComponent(userId)}`,
        { method: "DELETE" }
    );
    if (!response.ok) {
        throw new Error(await rejection(response, "Couldn't delete the transaction."));
    }
}

/** Spring's GlobalExceptionHandler answers 400s with { message }: a reason the user can read. */
async function rejection(response: Response, fallback: string): Promise<string> {
    const body = await response.json().catch(() => null);
    return response.status === 400 && body?.message ? String(body.message).replace(/^invalid transaction: /, "") : fallback;
}


/**
 * Get a specific transaction.
 *
 * GET /api/transactions/{id}
 */
export async function getTransaction(
    transactionId: string
): Promise<Transaction> {

    const response = await fetch(
        `${API_URL}/api/transactions/${transactionId}`
    );

    if (!response.ok) {
        const errorText = await response.text();

        throw new Error(
            `Failed to fetch transaction: ${response.status}: ${errorText}`
        );
    }

    return response.json();
}