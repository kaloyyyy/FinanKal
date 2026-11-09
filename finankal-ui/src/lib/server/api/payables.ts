import { api } from "./client";
import type { Payable } from "$lib/utils/payables";

/** GET /api/users/{userId}/payables: unpaid loan and card installments, soonest first. */
export async function getPayables(userId: string): Promise<Payable[]> {
    const payables = await api<Payable[]>(`/api/users/${userId}/payables`);
    return payables.map((p) => ({ ...p, amount: Number(p.amount) }));
}

export interface NewInstallments {
    name: string;
    sourceAccountId: string;
    accountId: string;
    installments: { dueDate: string; amount: string }[]; // yyyy-MM-dd, decimal string; in due-date order
}

/** POST /api/users/{userId}/payables: records a loan's installments. Throws with the API's error body when rejected. */
export async function createInstallments(userId: string, schedule: NewInstallments): Promise<{ scheduleId: string }> {
    return api(`/api/users/${userId}/payables`, { method: "POST", body: JSON.stringify(schedule) });
}
