import { api } from './client';
import type {LedgerEntryDto} from "$lib/types/ledger";

export interface LoanDetails {
    accountId: string;
    name: string;
    bookedOn: string;
    dailyRate: number;
    principalOn: string;
    principal: number;
}

export interface Account {
    id: string;
    name: string;
    type: string;
    balance: number;
    creditLimit?: number;
    billingDay?: number;
    statementShift?: string;
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

export async function getLoanDetails(userId: string, accountId: string): Promise<LoanDetails> {
    return api<LoanDetails>(`/api/users/${userId}/loans/${accountId}`);
}
export async function updateLoanDetails(userId: string, details: LoanDetails): Promise<void> {
    await api<void>(`/api/users/${userId}/loans/${details.accountId}`, { method: "PUT", body: JSON.stringify(details) });
}
export async function deleteLoan(userId: string, accountId: string): Promise<void> {
    await api<void>(`/api/users/${userId}/loans/${accountId}`, { method: "DELETE" });
}
