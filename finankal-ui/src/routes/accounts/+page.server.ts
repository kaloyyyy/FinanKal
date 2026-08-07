import type { PageServerLoad } from './$types';
import { getUserAccounts } from '$lib/server/api/accounts';

// TODO: Replace with authenticated user id
const USER_ID = 'bcc539f6-fc94-4a2e-9ce6-3c03c0560204';

export const load: PageServerLoad = async () => {
    return {
        accounts: await getUserAccounts(USER_ID)
    };
};