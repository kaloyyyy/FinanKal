import type { PageServerLoad } from "./$types";

import { getUserAccounts } from "$lib/server/api/accounts";
import { getMoneyFlows } from "$lib/server/api/flows";

const USER_ID = "bcc539f6-fc94-4a2e-9ce6-3c03c0560204";

export const load: PageServerLoad = async ({ url }) => {
    // ?month=2026-10 limits to one month; no month = all time.
    const month = /^\d{4}-\d{2}$/.test(url.searchParams.get("month") ?? "") ? url.searchParams.get("month")! : "";
    const [y, m] = month.split("-").map(Number);
    const from = month ? `${month}-01` : undefined;
    const to = month ? new Date(Date.UTC(y, m, 0)).toISOString().slice(0, 10) : undefined; // last day of the month

    const [accounts, flows] = await Promise.all([getUserAccounts(USER_ID), getMoneyFlows(USER_ID, from, to)]);

    // The last 12 months, newest first, for the picker (Philippine calendar).
    const now = new Date(Date.now() + 8 * 60 * 60 * 1000);
    const months = Array.from({ length: 12 }, (_, i) =>
        new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth() - i, 1)).toISOString().slice(0, 7)
    );

    return { accounts, flows, month, months };
};
