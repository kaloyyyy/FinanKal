import { formatDate } from "./date";

/** yyyy-MM-dd calendar days, as the API sends them. */
export interface StatementPeriod {
    accountId: string;
    bank: string;
    periodStart: string;
    periodEnd: string;
}

export type Coverage = "full" | "partial" | "missing";

export interface ChecklistCell {
    month: string; // yyyy-MM
    coverage: Coverage;
    /** The statement periods that touch this month, for a tooltip. */
    periods: string[];
}

export interface ChecklistRow {
    accountId: string;
    accountName: string;
    accountType: string;
    bank: string | null; // code of the bank its statements came from
    cells: ChecklistCell[];
}

const DAY = 24 * 60 * 60 * 1000;
const day = (iso: string) => Date.parse(`${iso}T00:00:00Z`) / DAY;
const monthKey = (d: Date) => d.toISOString().slice(0, 7);

/**
 * Which months each account has statements for, from its first statement (or the most recent `span`
 * months, whichever is shorter) up to this month. A month is full when its statements cover every day,
 * partial when they cover some.
 *
 * A credit card bills in cycles that cross months (24 Sep – 23 Oct), so for a card a month is full when
 * the bill closing in it is on file, and missing otherwise.
 */
export function statementChecklist(
    accounts: { id: string; name: string; type?: string }[],
    statements: StatementPeriod[],
    today: Date,
    span = 12
): { months: string[]; rows: ChecklistRow[] } {
    const thisMonth = new Date(Date.UTC(today.getUTCFullYear(), today.getUTCMonth(), 1));
    const oldestShown = new Date(Date.UTC(thisMonth.getUTCFullYear(), thisMonth.getUTCMonth() - (span - 1), 1));
    const earliest = statements.length
        ? new Date(Math.min(...statements.map((s) => day(s.periodStart))) * DAY)
        : thisMonth;
    const first = earliest > oldestShown ? new Date(Date.UTC(earliest.getUTCFullYear(), earliest.getUTCMonth(), 1)) : oldestShown;

    const months: string[] = [];
    for (let m = first; m <= thisMonth; m = new Date(Date.UTC(m.getUTCFullYear(), m.getUTCMonth() + 1, 1))) {
        months.push(monthKey(m));
    }

    const rows = accounts.map((account) => {
        const own = statements.filter((s) => s.accountId === account.id);
        const card = account.type === "CREDIT_CARD";
        const cells = months.map((month): ChecklistCell => {
            if (card) {
                // A monthly bill (a cycle of about a month) ticks its month; a shorter transaction history only part of it.
                const ending = own.filter((s) => s.periodEnd.startsWith(month));
                const isBill = (s: StatementPeriod) => day(s.periodEnd) - day(s.periodStart) >= 27;
                return {
                    month,
                    coverage: ending.some(isBill) ? "full" : ending.length ? "partial" : "missing",
                    periods: ending.map((s) => `${formatDate(s.periodStart)} – ${formatDate(s.periodEnd)}`)
                };
            }
            const start = day(`${month}-01`);
            const [y, m] = month.split("-").map(Number);
            const end = day(new Date(Date.UTC(y, m, 0)).toISOString().slice(0, 10)); // last day of the month

            const covered = new Set<number>();
            const periods: string[] = [];
            for (const s of own) {
                const from = Math.max(start, day(s.periodStart));
                const to = Math.min(end, day(s.periodEnd));
                if (from > to) continue;
                periods.push(`${formatDate(s.periodStart)} – ${formatDate(s.periodEnd)}`);
                for (let d = from; d <= to; d++) covered.add(d);
            }
            const coverage: Coverage =
                covered.size === 0 ? "missing" : covered.size === end - start + 1 ? "full" : "partial";
            return { month, coverage, periods };
        });
        return {
            accountId: account.id,
            accountName: account.name,
            accountType: account.type ?? "ASSET",
            bank: own.at(-1)?.bank ?? null,
            cells
        };
    });

    return { months, rows };
}
