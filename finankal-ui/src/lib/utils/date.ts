const parts = new Intl.DateTimeFormat("en-US", {
    month: "short",
    day: "2-digit",
    year: "numeric",
    timeZone: "Asia/Manila"
});

/** The yyyy-MM-dd a card's statement closes in a month (0-based; overflow rolls into the next year), moved off weekends per its shift. */
export function statementDate(year: number, month: number, billingDay: number, shift?: string): string {
    const last = new Date(Date.UTC(year, month + 1, 0)).getUTCDate();
    const close = new Date(Date.UTC(year, month, Math.min(Number(billingDay), last)));
    if (shift === "NEXT_BUSINESS_DAY") {
        while (close.getUTCDay() === 0 || close.getUTCDay() === 6) close.setUTCDate(close.getUTCDate() + 1);
    } else if (shift === "NEAREST_BUSINESS_DAY") {
        if (close.getUTCDay() === 6) close.setUTCDate(close.getUTCDate() - 1);
        else if (close.getUTCDay() === 0) close.setUTCDate(close.getUTCDate() + 1);
    }
    return close.toISOString().slice(0, 10);
}

/** The card's next statement date on or after today (Philippine calendar). */
export function nextStatementDate(billingDay: number, shift?: string, now = new Date()): string {
    const today = new Date(now.getTime() + 8 * 60 * 60 * 1000).toISOString().slice(0, 10);
    const [year, month] = today.split("-").map(Number);
    const close = statementDate(year, month - 1, billingDay, shift);
    return close >= today ? close : statementDate(year, month, billingDay, shift);
}

/** "NOV 2026", the month of formatDate's date. */
export function formatMonth(value: string | Date | null | undefined): string {
    const date = formatDate(value);
    return date === "—" ? date : `${date.slice(0, 3)} ${date.slice(7)}`;
}

/**
 * "NOV 09 2026", on the Philippine calendar. Takes a timestamp or a yyyy-MM-dd calendar day;
 * "—" when there's no valid date.
 */
export function formatDate(value: string | Date | null | undefined): string {
    if (!value) return "—";
    const date =
        typeof value === "string" && /^\d{4}-\d{2}-\d{2}$/.test(value)
            ? new Date(`${value}T00:00:00+08:00`) // a calendar day, not midnight UTC
            : new Date(value);
    if (Number.isNaN(date.getTime())) return "—";

    const p = Object.fromEntries(parts.formatToParts(date).map((part) => [part.type, part.value]));
    return `${p.month.toUpperCase()} ${p.day} ${p.year}`;
}
