const parts = new Intl.DateTimeFormat("en-US", {
    month: "short",
    day: "2-digit",
    year: "numeric",
    timeZone: "Asia/Manila"
});

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
