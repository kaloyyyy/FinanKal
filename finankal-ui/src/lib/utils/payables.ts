/** One unpaid installment, as GET /api/users/{id}/payables returns it. dueDate is yyyy-MM-dd. */
export interface Payable {
    name: string;
    accountId: string;
    accountName: string;
    dueDate: string;
    scheduledDate?: string;
    amount: number;
    installmentNumber: number | null;
    installments: number | null;
    sourceAccountId?: string | null;
    scheduleId?: string | null;
    occurrenceId?: string | null;
    /** How much of the amount holds a card's limit until paid (installment principal); unset = all of it. */
    principal?: number | null;
    /** Paikot: set when a card is linked to pay this account's dues. Charge it from payFrom to payBy and the due moves to dueDate. */
    paikot?: { cardId: string; cardName: string; payFrom: string; payBy: string; dueDate: string; fee: number } | null;
}

/** Each due where it lands once its linked card pays it: on the card's next due date, with the fee added. */
export const applyPaikot = (payables: Payable[]): Payable[] =>
    payables.map((p) =>
        p.paikot
            ? { ...p, dueDate: p.paikot.dueDate, amount: Math.round((Number(p.amount) + Number(p.paikot.fee)) * 100) / 100 }
            : p
    );

/** A card's spending limit now and as its dues get paid. Amounts are what's left to spend. */
export interface CardLimit {
    id: string;
    name: string;
    limit: number;
    now: number;
    afterThisMonth: number; // once this month's dues on the card are paid
    afterNextMonth: number; // and next month's
    afterPaikot: number; // and after the paikot charges put on it in that time
    short: { amount: number; date: string } | null; // the first paikot charge that doesn't fit
}

/**
 * What the card owes and its unpaid installments hold the limit: their principal when the plan has a monthly
 * rate, like the bank counts it, else the whole installment. Each due paid frees what it held (on its due
 * date, or when another card pays it), and a paikot charge on the card takes the amount plus the fee.
 */
export function cardLimits(
    accounts: { id: string; name: string; type: string; balance: number; creditLimit?: number }[],
    payables: Payable[],
    thisMonth: string,
    nextMonth: string
): CardLimit[] {
    const cents = (n: number) => Math.round(Number(n) * 100);
    const held = (p: Payable) => cents(p.principal ?? p.amount);
    const total = (items: Payable[]) => items.reduce((c, p) => c + held(p), 0);
    const until = (p: Payable, month: string) => p.dueDate.slice(0, 7) <= month;
    return accounts
        .filter((a) => a.type === "CREDIT_CARD" && a.creditLimit != null)
        .map((card) => {
            const own = payables.filter((p) => p.accountId === card.id);
            // The card's dues are its bills (what it owes) plus its unpaid installments, some already on the current bill.
            const now = cents(card.creditLimit!) - Math.max(-cents(card.balance), total(own));
            const afterThisMonth = now + total(own.filter((p) => until(p, thisMonth)));
            const afterNextMonth = now + total(own.filter((p) => until(p, nextMonth)));
            // In date order, payments before charges on the same day: repay first, then borrow again.
            const events = [
                ...own.filter((p) => until(p, nextMonth)).map((p) => ({ date: p.paikot?.payBy ?? p.dueDate, cents: held(p) })),
                ...payables
                    .filter((p) => p.paikot?.cardId === card.id && until(p, nextMonth))
                    .map((p) => ({ date: p.paikot!.payBy, cents: -cents(p.amount) - cents(p.paikot!.fee) }))
            ].sort((a, b) => a.date.localeCompare(b.date) || b.cents - a.cents);
            let left = now;
            let short: CardLimit["short"] = null;
            for (const e of events) {
                left += e.cents;
                if (e.cents < 0 && left < 0 && !short) short = { amount: -left / 100, date: e.date };
            }
            return {
                id: card.id, name: card.name, limit: Number(card.creditLimit),
                now: now / 100, afterThisMonth: afterThisMonth / 100, afterNextMonth: afterNextMonth / 100,
                afterPaikot: left / 100, short
            };
        });
}

export interface PayableMonth {
    month: string; // yyyy-MM
    total: number;
    items: Payable[];
}

/**
 * What's owed in total, in the current month and in the next one (Philippine calendar).
 * Overdue installments from earlier months count as this month's: they're due now.
 */
export function summarizePayables(payables: Payable[], today: Date) {
    const thisMonth = monthKey(today);
    const nextMonth = monthKey(new Date(Date.UTC(today.getUTCFullYear(), today.getUTCMonth() + 1, 1)));
    const pick = (match: (month: string) => boolean): PayableMonth => {
        const items = payables.filter((p) => match(p.dueDate.slice(0, 7)));
        return { month: "", total: sum(items), items };
    };
    return {
        total: sum(payables),
        thisMonth: { ...pick((m) => m <= thisMonth), month: thisMonth },
        nextMonth: { ...pick((m) => m === nextMonth), month: nextMonth }
    };
}

export interface NetPositions {
    money: number; // in your bank and wallet accounts
    debt: number; // owed on loan and card accounts
    salary: number; // predicted for next month: average monthly salary
    salaryMonths: number; // how many months that average is over (0 = no salary history)
    dueThisMonth: number;
    dueNextMonth: number;
    net: number; // money − debt
    thisMonth: number; // money − due this month
    nextMonth: number; // money + salary − due next month
}

/**
 * Money is what bank and wallet accounts hold; debt is what loan and card accounts owe (balances are
 * debits − credits, so an amount owed is negative). The average salary is over months that had any.
 */
export function netPositions(
    accounts: { type: string; balance: number }[],
    monthlySalaries: number[],
    dueThisMonth: number,
    dueNextMonth: number
): NetPositions {
    const cents = (n: number) => Math.round(Number(n) * 100);
    const money = accounts.filter((a) => a.type === "ASSET").reduce((c, a) => c + cents(a.balance), 0);
    const debt = accounts
        .filter((a) => a.type === "LIABILITY" || a.type === "CREDIT_CARD")
        .reduce((c, a) => c + Math.max(0, -cents(a.balance)), 0);
    const paid = monthlySalaries.filter((s) => s > 0);
    const salary = paid.length ? Math.round(paid.reduce((c, s) => c + cents(s), 0) / paid.length) : 0;
    const thisDue = cents(dueThisMonth);
    const nextDue = cents(dueNextMonth);
    return {
        money: money / 100,
        debt: debt / 100,
        salary: salary / 100,
        salaryMonths: paid.length,
        dueThisMonth: thisDue / 100,
        dueNextMonth: nextDue / 100,
        net: (money - debt) / 100,
        thisMonth: (money - thisDue) / 100,
        nextMonth: (money + salary - nextDue) / 100
    };
}

export interface PositionTile {
    label: string;
    hint: string;
    question: string; // what this position answers, in plain words
    when: string; // the period it covers
    value: number;
    terms: { sign: "" | "+" | "−"; label: string; amount: number }[];
}

/** The three net positions as labelled sums, so each result can be checked at a glance. */
export function positionTiles(positions: NetPositions, thisMonth: string, nextMonth: string, monthLabel: (m: string) => string): PositionTile[] {
    const salaryHint = positions.salaryMonths
        ? `With your average salary over the last ${positions.salaryMonths} ${positions.salaryMonths === 1 ? "month" : "months"}`
        : "No salary history yet, so none is predicted";
    return [
        {
            label: "Net position",
            hint: "Everything you have, minus everything you owe",
            question: "If you paid every debt today",
            when: "All loans and cards",
            value: positions.net,
            terms: [
                { sign: "", label: "You have", amount: positions.money },
                { sign: "−", label: "You owe", amount: positions.debt }
            ]
        },
        {
            label: `This month · ${monthLabel(thisMonth)}`,
            hint: "After this month's installments",
            question: "After this month's bills",
            when: monthLabel(thisMonth),
            value: positions.thisMonth,
            terms: [
                { sign: "", label: "You have", amount: positions.money },
                { sign: "−", label: "Due this month", amount: positions.dueThisMonth }
            ]
        },
        {
            label: `Next month · ${monthLabel(nextMonth)}`,
            hint: salaryHint,
            question: "After next month's bills",
            when: positions.salaryMonths ? `${monthLabel(nextMonth)}, with your usual salary` : monthLabel(nextMonth),
            value: positions.nextMonth,
            terms: [
                { sign: "", label: "You have", amount: positions.money },
                { sign: "+", label: "Predicted salary", amount: positions.salary },
                { sign: "−", label: "Due next month", amount: positions.dueNextMonth }
            ]
        }
    ];
}

/** One account's share of a term in a position's split bar. */
export interface Slice {
    id: string;
    name: string;
    amount: number;
    predicted?: boolean; // not in the account yet (next month's salary)
    label?: string; // shown in the legend instead of name (name still picks the bank colour)
}

/** Installments grouped by the account they pay down, biggest first, as bar slices. */
export function slicesByAccount(items: Payable[]): Slice[] {
    const byAccount = new Map<string, Slice>();
    for (const p of items) {
        const slice = byAccount.get(p.accountId) ?? { id: `due-${p.accountId}`, name: p.accountName, amount: 0 };
        slice.amount = Math.round((slice.amount + Number(p.amount)) * 100) / 100;
        byAccount.set(p.accountId, slice);
    }
    return [...byAccount.values()].sort((a, b) => b.amount - a.amount);
}

/**
 * How much of what's going out the money coming in covers, for a position's meter.
 * fill is 0–1; text says it in words. Small shares keep one decimal ("1.5%") so they don't read as 0%.
 */
export function coverage(tile: PositionTile): { fill: number; covered: boolean; text: string } {
    const cents = (n: number) => Math.round(n * 100);
    const have = tile.terms.filter((t) => t.sign !== "−").reduce((c, t) => c + cents(t.amount), 0);
    const owe = tile.terms.filter((t) => t.sign === "−").reduce((c, t) => c + cents(t.amount), 0);
    if (owe <= 0) return { fill: 1, covered: true, text: "Nothing due" };
    if (have >= owe) return { fill: 1, covered: true, text: `Covered, with ${money(have - owe)} to spare` };
    const share = (Math.max(have, 0) / owe) * 100;
    const pct = share < 10 ? Math.round(share * 10) / 10 : Math.round(share);
    return { fill: Math.max(have, 0) / owe, covered: false, text: `Covers ${pct}% of it, ${money(owe - have)} short` };
}

const money = (cents: number) =>
    new Intl.NumberFormat("en-PH", { style: "currency", currency: "PHP" }).format(cents / 100);

// Sum in centavos so 943.37 + 3723.50 is exactly 4666.87.
const sum = (items: Payable[]) => items.reduce((cents, p) => cents + Math.round(Number(p.amount) * 100), 0) / 100;

const monthKey = (d: Date) => d.toISOString().slice(0, 7);
