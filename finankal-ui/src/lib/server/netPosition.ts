import { getMoneyFlows } from "$lib/server/api/flows";
import { getPayables } from "$lib/server/api/payables";
import { applyPaikot, cardLimits, netPositions, summarizePayables } from "$lib/utils/payables";

// The income account statement imports file payroll under (BankStatementParser.category).
const SALARY_ACCOUNT = "Salary Income";
const SALARY_MONTHS = 3;

/**
 * Net position now, this month and next month, plus the installments behind them.
 * Null when payables can't be loaded, so a page can leave the section out instead of failing.
 */
export async function loadNetPositions(
    userId: string,
    accounts: { id: string; name: string; type: string; balance: number; creditLimit?: number }[]
) {
    // Philippine calendar, so a month starts at Manila midnight.
    const today = new Date(Date.now() + 8 * 60 * 60 * 1000);
    // The last few complete months (this month's salary isn't all in yet), as [first day, last day].
    const salaryMonths = Array.from({ length: SALARY_MONTHS }, (_, i) => {
        const first = new Date(Date.UTC(today.getUTCFullYear(), today.getUTCMonth() - 1 - i, 1));
        const last = new Date(Date.UTC(first.getUTCFullYear(), first.getUTCMonth() + 1, 0));
        return [first, last].map((d) => d.toISOString().slice(0, 10));
    });

    const [payables, salaryFlows] = await Promise.all([
        getPayables(userId).catch((error) => {
            console.error("Payables failed to load:", error);
            return null;
        }),
        Promise.all(salaryMonths.map(([from, to]) => getMoneyFlows(userId, from, to))).catch((error) => {
            console.error("Salary history failed to load:", error);
            return [];
        })
    ]);
    if (!payables) return null;

    const due = summarizePayables(payables, today);
    const salaryAccount = accounts.find((a) => a.type === "INCOME" && a.name === SALARY_ACCOUNT);
    const monthlySalaries = salaryFlows.map((flows) =>
        flows.filter((f) => f.fromAccountId === salaryAccount?.id).reduce((sum, f) => sum + f.amount, 0)
    );

    // The account the salary is paid into: the one that received the most of it.
    const into = new Map<string, number>();
    for (const flows of salaryFlows) {
        for (const f of flows) {
            if (f.fromAccountId === salaryAccount?.id) into.set(f.toAccountId, (into.get(f.toAccountId) ?? 0) + f.amount);
        }
    }
    const salaryAccountId = [...into.entries()].sort((a, b) => b[1] - a[1])[0]?.[0];
    const salaryInto = accounts.find((a) => a.id === salaryAccountId);

    return {
        thisMonth: due.thisMonth,
        nextMonth: due.nextMonth,
        // Next month again, with every linked due moved onto the card that pays it.
        paikot: summarizePayables(applyPaikot(payables), today).nextMonth,
        // Every due a link moves, soonest charge first: what to pay with which card, and when.
        paikotMoves: payables.filter((p) => p.paikot).sort((a, b) => a.paikot!.payBy.localeCompare(b.paikot!.payBy)),
        cardLimits: cardLimits(accounts, payables, due.thisMonth.month, due.nextMonth.month),
        positions: netPositions(accounts, monthlySalaries, due.thisMonth.total, due.nextMonth.total),
        salaryInto: salaryInto ? { id: salaryInto.id, name: salaryInto.name } : null
    };
}
