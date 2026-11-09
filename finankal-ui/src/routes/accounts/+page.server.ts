import type { PageServerLoad } from './$types';
import { getUserAccounts } from '$lib/server/api/accounts';
import { loadNetPositions } from '$lib/server/netPosition';

// TODO: Replace with authenticated user id
const USER_ID = 'bcc539f6-fc94-4a2e-9ce6-3c03c0560204';

export const load: PageServerLoad = async () => {
    const accounts = await getUserAccounts(USER_ID);
    return {
        accounts,
        netPosition: await loadNetPositions(USER_ID, accounts)
    };
};