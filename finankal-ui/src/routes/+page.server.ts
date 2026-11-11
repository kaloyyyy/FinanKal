import { fail } from "@sveltejs/kit";
import type { Actions, PageServerLoad } from "./$types";
import { setPaikotLink } from "$lib/server/api/payables";
import { getUserFinancialSummary } from "$lib/server/api/users";
import {getRecentTransactions} from "$lib/server/api/transactions";
import { getLedgerEntries, getUserAccounts } from "$lib/server/api/accounts";
import { formatMonth } from "$lib/utils/date";
import { loadNetPositions } from "$lib/server/netPosition";


const USER_ID =
    "bcc539f6-fc94-4a2e-9ce6-3c03c0560204";


export const load: PageServerLoad = async () => {

    const [summary, recentTransactions, accounts] = await Promise.all([
        getUserFinancialSummary(USER_ID),
        getRecentTransactions(USER_ID, 5),
        getUserAccounts(USER_ID)
    ]);

    const [netWorthHistory, netPosition] = await Promise.all([
        loadNetWorthHistory(accounts).catch((error) => {
            console.error("Net worth history failed to load:", error);
            return [];
        }),
        loadNetPositions(USER_ID, accounts)
    ]);
    return { summary, recentTransactions, accounts, netWorthHistory, userId: USER_ID, netPosition };
};

const MANILA = 8 * 60 * 60 * 1000;

/**
 * Net worth at the end of each of the last six months (this month: so far), like the engine's net worth:
 * debits − credits over what you have and what you owe. Months before the first entry are left out.
 * shortcut: reads every such account's whole ledger on each load; add an engine endpoint that sums by month if the dashboard gets slow.
 */
async function loadNetWorthHistory(accounts: { id: string; type: string }[]) {
    const ledgers = await Promise.all(
        accounts.filter((a) => ["ASSET", "LIABILITY", "CREDIT_CARD"].includes(a.type)).map((a) => getLedgerEntries(a.id))
    );
    // The change in each Philippine calendar month, in centavos.
    const changes = new Map<string, number>();
    for (const entry of ledgers.flat()) {
        const at = Date.parse(entry.createdAt ?? "");
        if (Number.isNaN(at)) continue;
        const month = new Date(at + MANILA).toISOString().slice(0, 7);
        changes.set(month, (changes.get(month) ?? 0) + (entry.type === "DEBIT" ? 1 : -1) * Math.round(Number(entry.amount) * 100));
    }
    const first = [...changes.keys()].sort()[0];
    const today = new Date(Date.now() + MANILA);
    return Array.from({ length: 6 }, (_, i) => new Date(Date.UTC(today.getUTCFullYear(), today.getUTCMonth() - 5 + i, 1)).toISOString().slice(0, 7))
        .filter((month) => first && month >= first)
        .map((month) => ({
            label: formatMonth(`${month}-01`),
            value: [...changes].reduce((cents, [m, change]) => (m <= month ? cents + change : cents), 0) / 100
        }));
}

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
