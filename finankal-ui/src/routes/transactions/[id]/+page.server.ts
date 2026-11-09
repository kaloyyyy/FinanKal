import { error, fail, redirect } from "@sveltejs/kit";
import { getUserAccounts } from "$lib/server/api/accounts";
import { deleteTransaction, getTransaction, updateTransaction } from "$lib/server/api/transactions";
import type { Actions, PageServerLoad } from "./$types";

const USER_ID = "bcc539f6-fc94-4a2e-9ce6-3c03c0560204";

/** Where "back" goes: the account ledger the user came from, else the transaction list. */
function backTo(url: URL): string {
    const from = url.searchParams.get("from");
    // Only this app's account pages, so the link can't send the user elsewhere.
    return from && /^\/accounts\/[\w-]+$/.test(from) ? from : "/transactions";
}

export const load: PageServerLoad = async ({ params, url }) => {
    try {
        const [transaction, accounts] = await Promise.all([
            getTransaction(params.id),
            getUserAccounts(USER_ID)
        ]);
        return {
            transaction,
            accounts: accounts.map(({ id, name, type }) => ({ id, name, type })),
            back: backTo(url)
        };
    } catch (e) {
        console.error(e);
        throw error(404, "Transaction not found");
    }
};

export const actions: Actions = {
    update: async ({ request, params }) => {
        const raw = (await request.formData()).get("transaction");
        if (typeof raw !== "string") {
            return fail(400, { error: "The changes didn't reach the server. Try again." });
        }
        try {
            await updateTransaction(USER_ID, params.id, JSON.parse(raw));
            return { saved: true };
        } catch (e) {
            console.error("Updating transaction failed:", e);
            return fail(400, { error: (e as Error).message });
        }
    },

    delete: async ({ params, url }) => {
        try {
            await deleteTransaction(USER_ID, params.id);
        } catch (e) {
            console.error("Deleting transaction failed:", e);
            return fail(400, { error: (e as Error).message });
        }
        throw redirect(303, backTo(url));
    }
};
