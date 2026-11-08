import { getUserTransactions } from "$lib/server/api/transactions";

export async function load() {
    const userId = "bcc539f6-fc94-4a2e-9ce6-3c03c0560204";

    const result = await getUserTransactions(userId);

    console.log("TRANSACTION RESULT:", result);
    console.log("TRANSACTION COUNT:", result.transactions.length);
    console.log("TOTAL:", result.total);

    return {
        transactions: result.transactions,
        total: result.total,
        limit: result.limit,
        offset: result.offset
    };
}