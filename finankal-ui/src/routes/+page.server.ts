import type { PageServerLoad } from "./$types";
import { getUserFinancialSummary } from "$lib/server/api/users";


const USER_ID =
    "bcc539f6-fc94-4a2e-9ce6-3c03c0560204";


export const load: PageServerLoad = async () => {

    const summary =
        await getUserFinancialSummary(USER_ID);

    return {
        summary
    };
};