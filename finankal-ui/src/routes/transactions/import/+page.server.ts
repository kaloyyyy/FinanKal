import { fail } from "@sveltejs/kit";
import type { Actions, PageServerLoad } from "./$types";

import { getUserAccounts } from "$lib/server/api/accounts";
import { BANKS, StatementImportError, getBankStatements, importBankStatements, type Bank } from "$lib/server/api/bankStatements";
import { statementChecklist } from "$lib/utils/statementChecklist";

const USER_ID = "bcc539f6-fc94-4a2e-9ce6-3c03c0560204";

export const load: PageServerLoad = async () => {
    const [accounts, statements] = await Promise.all([getUserAccounts(USER_ID), getBankStatements(USER_ID)]);
    // Statements go into bank, wallet and credit card accounts.
    const sourceAccounts = accounts.filter((a) => a.type === "ASSET" || a.type === "CREDIT_CARD");

    return {
        banks: BANKS,
        sourceAccounts,
        // Today in Manila, so a new month starts at Philippine midnight.
        checklist: statementChecklist(sourceAccounts, statements, new Date(Date.now() + 8 * 60 * 60 * 1000))
    };
};

export const actions: Actions = {
    import: async ({ request }) => {
        const formData = await request.formData();

        const files = formData.getAll("files").filter((f): f is File => f instanceof File && f.size > 0);
        const bank = String(formData.get("bank") ?? ""); // "" = detect from each statement
        const sourceAccountId = String(formData.get("sourceAccountId") ?? ""); // "" = find by account number
        const password = String(formData.get("password") ?? "");

        if (bank && !BANKS.some((b) => b.value === bank)) {
            return fail(400, { error: "Choose a bank from the list, or let FinanKal detect it." });
        }
        if (!files.length) {
            return fail(400, { error: "Add the statement PDFs you downloaded from your bank." });
        }

        try {
            const results = await importBankStatements({
                files,
                userId: USER_ID,
                bank: (bank || undefined) as Bank | undefined,
                sourceAccountId: sourceAccountId || undefined,
                password
            });
            return { success: true, results };
        } catch (error) {
            console.error("Statement import failed:", error);

            if (error instanceof StatementImportError && error.status === 400) {
                return fail(400, { error: error.message });
            }
            return fail(500, { error: "The statements couldn't be imported. Check that the API is running and try again." });
        }
    }
};
