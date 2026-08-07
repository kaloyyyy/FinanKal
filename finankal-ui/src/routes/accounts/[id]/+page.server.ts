import { error } from "@sveltejs/kit";
import { getAccount, getLedgerEntries } from "$lib/server/api/accounts";
import type { PageServerLoad } from "./$types";

export const load: PageServerLoad = async ({ params }) => {
    try {
        const [account, entries] = await Promise.all([
            getAccount(params.id),
            getLedgerEntries(params.id)
        ]);

        return {
            account,
            entries
        };
    } catch (e) {
        console.error(e);
        throw error(404, "Account not found");
    }
};