import { getUserTransactions } from "$lib/server/api/transactions";

export async function load() {
    const userId = "bcc539f6-fc94-4a2e-9ce6-3c03c0560204";

    const transactions = await getUserTransactions(userId);

    return {
        transactions
    };
}