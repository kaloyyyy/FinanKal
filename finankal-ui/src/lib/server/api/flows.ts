import { api } from "./client";
import type { Flow } from "$lib/utils/moneyFlowGraph";

/**
 * GET /api/users/{userId}/money-flows?from=&to= (yyyy-MM-dd, inclusive, both optional)
 * Money moved from one account to another, largest first.
 */
export async function getMoneyFlows(userId: string, from?: string, to?: string): Promise<Flow[]> {
    const query = new URLSearchParams();
    if (from) query.set("from", from);
    if (to) query.set("to", to);
    const flows = await api<Flow[]>(`/api/users/${userId}/money-flows${query.size ? `?${query}` : ""}`);
    // BigDecimal arrives as a JSON number or string depending on Jackson settings.
    return flows.map((f) => ({ ...f, amount: Number(f.amount) }));
}
