const API_URL = "http://localhost:8080";

/**
 * Create a new double-entry transaction.
 */
export async function createTransaction(payload: {
    description: string;
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
    accountId: string;
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