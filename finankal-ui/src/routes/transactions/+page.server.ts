import { fail } from "@sveltejs/kit";
import type { Actions } from "./$types";

import { clearUserTransactions, getUserTransactions } from "$lib/server/api/transactions";
import { getUserAccounts } from "$lib/server/api/accounts";

const userId = "bcc539f6-fc94-4a2e-9ce6-3c03c0560204";

export async function load() {
    // Accounts let the page show names instead of account ids on each entry.
    const [result, accounts] = await Promise.all([
        getUserTransactions(userId),
        getUserAccounts(userId)
    ]);

    return {
        transactions: result.transactions,
        total: result.total,
        accounts
    };
}

export const actions: Actions = {
    clear: async () => {
        try {
            const { deletedTransactions } = await clearUserTransactions(userId);
            return { cleared: deletedTransactions };
        } catch (error) {
            console.error("Clearing transactions failed:", error);
            return fail(500, { clearError: "Transactions couldn't be cleared. Check that the API is running and try again." });
        }
    }
};
