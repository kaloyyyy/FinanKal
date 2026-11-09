/**
 * Brand colours for account slices, matched on the account's name.
 *
 * Patterns are 2px lines. The owed (dark) and predicted (light) hatches are soft, at 45° every 6px.
 * A bank's own lines are solid colour, so they're spaced every 10px to look as light. MariBank and
 * UnionBank are both orange, so MariBank carries blue lines; Mari Loan cross-hatches blue and red.
 * shortcut: matched by name; move to a per-account colour once accounts can be edited in the app.
 */
const BANK_PERIOD = 10;
const BANK_LINE_OFFSET = 3; // between the overlay hatch's lines, so neither covers the other

/** One layer of 2px lines in `color`, at `angle`, one every BANK_PERIOD px. */
const lines = (color: string, angle: number) =>
    `repeating-linear-gradient(${angle}deg, transparent ${BANK_LINE_OFFSET}px, ${color} ${BANK_LINE_OFFSET}px ${BANK_LINE_OFFSET + 2}px, transparent ${BANK_LINE_OFFSET + 2}px ${BANK_PERIOD}px)`;

interface Brand {
    match: RegExp;
    base: string;
    lines?: { color: string; angle: number }[];
    marksDebt?: boolean;
    mark: string; // short letters shown on the account badge
    ink: string; // text colour that reads on `base`
    wallet?: boolean; // an e-wallet rather than a bank
}

const WHITE = "#ffffff";
const BRANDS: Brand[] = [
    // Mari Loan: orange cross-hatched blue one way, red the other. The red already says "owed", so no dark hatch on top.
    {
        match: /mari\s*loan|maribank\s*loan/i,
        base: "#f26b21",
        lines: [{ color: "#2563eb", angle: 45 }, { color: "#dc2626", angle: 135 }],
        marksDebt: true,
        mark: "MB",
        ink: WHITE
    },
    { match: /maribank|maricard/i, base: "#f26b21", lines: [{ color: "#2563eb", angle: 45 }], mark: "MB", ink: WHITE },
    { match: /unionbank/i, base: "#f7941d", mark: "UB", ink: WHITE },
    { match: /maya\s*savings/i, base: "#00c853", mark: "MY", ink: WHITE }, // Maya Bank, not the wallet
    { match: /maya/i, base: "#00c853", mark: "MY", ink: WHITE, wallet: true },
    { match: /gcash/i, base: "#007dfe", mark: "GC", ink: WHITE, wallet: true },
    { match: /\bbpi\b/i, base: "#b5121b", mark: "BPI", ink: WHITE },
    { match: /\bbdo\b/i, base: "#0033a0", mark: "BDO", ink: WHITE },
    { match: /gotyme/i, base: "#00a6a0", mark: "GT", ink: WHITE },
    { match: /atome/i, base: "#f5e100", mark: "AT", ink: "#1c1917" },
    { match: /spaylater|shopee/i, base: "#ee4d2d", mark: "SP", ink: WHITE }
];

/** What an account's badge shows: its bank's colour, letters and ink, or null when the bank isn't known. */
export function bankBadge(name: string): { background: string; mark: string; ink: string; wallet: boolean } | null {
    const brand = BRANDS.find((b) => b.match.test(name));
    if (!brand) return null;
    return { background: bankBackground(name)!, mark: brand.mark, ink: brand.ink, wallet: !!brand.wallet };
}

// The owed and predicted hatches, exactly as first designed (45°, lines at 0–2px of every 6px).
const OWED_HATCH = "repeating-linear-gradient(45deg, rgba(2, 6, 23, 0.6) 0 2px, transparent 2px 6px)";
const PREDICTED_HATCH = "repeating-linear-gradient(45deg, rgba(255, 255, 255, 0.45) 0 2px, transparent 2px 6px)";

/**
 * CSS background for an account's slice, or null when the bank isn't known (use the default colours).
 * "owed" adds the dark hatch (unless the brand already marks debt); "predicted" adds the light one.
 */
export function bankBackground(name: string, mode: "held" | "owed" | "predicted" = "held"): string | null {
    const brand = BRANDS.find((b) => b.match.test(name));
    if (!brand) return null;
    const overlay = mode === "predicted" ? PREDICTED_HATCH : mode === "owed" && !brand.marksDebt ? OWED_HATCH : null;
    const layers = [overlay, ...(brand.lines ?? []).map((l) => lines(l.color, l.angle))].filter(Boolean);
    return layers.length ? `${layers.join(", ")}, linear-gradient(${brand.base}, ${brand.base})` : brand.base;
}
