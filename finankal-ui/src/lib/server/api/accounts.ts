import { api } from './client';
import type {LedgerEntryDto} from "$lib/types/ledger";

export interface Account {
    id: string;
    name: string;
    type: string;
    balance: number;
    createdAt: string;
}

/**
 * GET /api/accounts/{id}
 */
export async function getAccount(id: string): Promise<Account> {
    return api<Account>(`/api/accounts/${id}`);
}

/**
 * GET /api/accounts/{id}/balance
 */
export async function getAccountBalance(id: string): Promise<number> {
    return api<number>(`/api/accounts/${id}/balance`);
}

export function getLedgerEntries(id: string) {

    return api<LedgerEntryDto[]>(
        `/api/ledger-entries/${id}`
    );
}

/**
 * GET /api/users/{userId}/accounts
 */
export async function getUserAccounts(userId: string): Promise<Account[]> {
    return api<Account[]>(`/api/users/${userId}/accounts`);
}