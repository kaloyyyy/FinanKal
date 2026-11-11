import { fail } from "@sveltejs/kit";
import type { Actions, PageServerLoad } from "./$types";
import { setPaikotLink } from "$lib/server/api/payables";
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

export const actions: Actions = {
    /** Links an account's dues to the card that can pay them (paikot); without a card, removes the link. */
    setPaikot: async ({ request }) => {
        const fields = await request.formData();
        const accountId = fields.get("accountId");
        const cardId = fields.get("cardId");
        if (typeof accountId !== "string" || !accountId) return fail(400, { paikotError: "Choose the loan or card to pay." });
        try {
            await setPaikotLink(USER_ID, accountId, typeof cardId === "string" && cardId ? cardId : null, Number(fields.get("feePercent")) || 0);
        } catch (e) {
            console.error("Saving the paikot link failed:", e);
            return fail(400, { paikotError: "Couldn't save the link. Pay a loan or card with a different card, at a fee of 0 to 100%." });
        }
    }
};
