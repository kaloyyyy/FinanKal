import { fail } from "@sveltejs/kit";
import type { Actions, PageServerLoad } from "./$types";

import { getUserAccounts } from "$lib/server/api/accounts";
import { createTransaction } from "$lib/server/api/transactions";

const USER_ID = "bcc539f6-fc94-4a2e-9ce6-3c03c0560204";

export const load: PageServerLoad = async () => {
    const accounts = await getUserAccounts(USER_ID);

    return {
        accounts
    };
};

export const actions: Actions = {
    create: async ({ request }) => {

        const formData = await request.formData();

        const transactionJson =
            formData.get("transaction");

        if (typeof transactionJson !== "string") {
            return fail(400, {
                error: "Missing transaction data"
            });
        }

        try {

            const payload = JSON.parse(transactionJson);

            console.log(
                "Creating transaction:",
                payload
            );

            const transaction =
                await createTransaction(payload);

            return {
                success: true,
                transaction
            };

        } catch (error) {

            console.error(
                "Transaction creation failed:",
                error
            );

            return fail(500, {
                error: "Failed to create transaction"
            });
        }
    }
};