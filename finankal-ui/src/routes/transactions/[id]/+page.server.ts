import { getTransaction } from "$lib/server/api/transactions";

export async function load({ params }) {
    const transaction = await getTransaction(params.id);

    return {
        transaction
    };
}