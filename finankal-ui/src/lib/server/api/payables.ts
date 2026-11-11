import { api } from "./client";
import type { Payable } from "$lib/utils/payables";

/** GET /api/users/{userId}/payables: unpaid loan and card installments, soonest first. */
export async function getPayables(userId: string): Promise<Payable[]> {
    const payables = await api<Payable[]>(`/api/users/${userId}/payables`);
    return payables.map((p) => ({ ...p, amount: Number(p.amount) }));
}

/** PUT /api/users/{userId}/paikot/{accountId}: the card that can pay this account's dues; a null cardId removes the link. */
export async function setPaikotLink(userId: string, accountId: string, cardId: string | null, feePercent = 0): Promise<void> {
    await api(`/api/users/${userId}/paikot/${accountId}`, { method: "PUT", body: JSON.stringify({ cardId, feePercent }) });
}

export interface NewInstallments {
    name: string;
    sourceAccountId: string;
    accountId: string;
    firstInstallmentNumber?: number;
    installments: { dueDate: string; amount: string }[]; // yyyy-MM-dd, decimal string; in due-date order
}

/** POST /api/users/{userId}/payables: records a loan's installments. Throws with the API's error body when rejected. */
export async function createInstallments(userId: string, schedule: NewInstallments): Promise<{ scheduleId: string }> {
    return api(`/api/users/${userId}/payables`, { method: "POST", body: JSON.stringify(schedule) });
}

export async function updateInstallments(userId: string, scheduleId: string, schedule: NewInstallments): Promise<void> {
    await api(`/api/users/${userId}/payables/${scheduleId}`, { method: "PUT", body: JSON.stringify(schedule) });
}
export async function deleteInstallments(userId: string, scheduleId: string): Promise<void> {
    await api(`/api/users/${userId}/payables/${scheduleId}`, { method: "DELETE" });
}
export async function payCardInstallment(userId: string, occurrenceId: string): Promise<void> {
    await api(`/api/users/${userId}/payables/occurrences/${occurrenceId}/pay`, { method: "POST" });
}
