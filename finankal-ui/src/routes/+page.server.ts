import type { PageServerLoad } from "./$types";
import { getUserFinancialSummary } from "$lib/server/api/users";
import {getRecentTransactions} from "$lib/server/api/transactions";
import { getUserAccounts } from "$lib/server/api/accounts";
import { loadNetPositions } from "$lib/server/netPosition";


const USER_ID =
    "bcc539f6-fc94-4a2e-9ce6-3c03c0560204";


export const load: PageServerLoad = async () => {

    const [summary, recentTransactions, accounts] = await Promise.all([
        getUserFinancialSummary(USER_ID),
        getRecentTransactions(USER_ID, 5),
        getUserAccounts(USER_ID)
    ]);

    // shortcut: no net-worth history endpoint yet, so the chart shows its empty state; fill from the API once it exists.
    const netWorthHistory: { label: string; value: number }[] = [];
    return {
        summary, recentTransactions, accounts, netWorthHistory, userId: USER_ID,
        netPosition: await loadNetPositions(USER_ID, accounts)
    };
};
