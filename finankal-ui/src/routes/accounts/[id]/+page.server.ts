import { error, fail } from "@sveltejs/kit";
import { getAccount, getLedgerEntries, getUserAccounts } from "$lib/server/api/accounts";
import { createInstallments, getPayables } from "$lib/server/api/payables";
import type { Actions, PageServerLoad } from "./$types";

const USER_ID = "bcc539f6-fc94-4a2e-9ce6-3c03c0560204";

export const load: PageServerLoad = async ({ params }) => {
    try {
        const [account, entries, payables, accounts] = await Promise.all([
            getAccount(params.id),
            getLedgerEntries(params.id),
            // The page still loads without deadlines if they can't be fetched.
            getPayables(USER_ID).catch((e) => {
                console.error("Payables failed to load:", e);
                return [];
            }),
            // Only needed to pick where installments are paid from.
            getUserAccounts(USER_ID).catch((e) => {
                console.error("Accounts failed to load:", e);
                return [];
            })
        ]);

        return {
            account,
            entries,
            // This account's own deadlines (a loan's installments, a card's bills), soonest first.
            deadlines: payables.filter((p) => p.accountId === params.id),
            // Installments are paid from a bank or wallet.
            payFrom: accounts.filter((a) => a.type === "ASSET").map(({ id, name }) => ({ id, name }))
        };
    } catch (e) {
        console.error(e);
        throw error(404, "Account not found");
    }
};

export const actions: Actions = {
    addInstallments: async ({ request, params }) => {
        const raw = (await request.formData()).get("schedule");
        if (typeof raw !== "string") {
            return fail(400, { error: "The schedule didn't reach the server. Try again." });
        }

        let schedule;
        try {
            schedule = JSON.parse(raw);
        } catch {
            return fail(400, { error: "The schedule didn't reach the server. Try again." });
        }

        try {
            await createInstallments(USER_ID, { ...schedule, accountId: params.id });
            return { added: schedule.installments.length as number };
        } catch (e) {
            console.error("Adding installments failed:", e);
            return fail(400, { error: reason(e) });
        }
    }
};

/** The engine's reason for a rejected schedule (the API answers 400 with it), or a generic one. */
function reason(e: unknown): string {
    try {
        const body = JSON.parse((e as Error).message);
        if (body.status === 400 && body.message) return String(body.message).replace(/^invalid schedule: /, "");
    } catch {
        // not a JSON error body
    }
    return "Couldn't save the installments. Check that the API and engine are running.";
}
