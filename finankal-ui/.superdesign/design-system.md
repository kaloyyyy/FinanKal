# FinanKal Design System

## Product context
FinanKal is a personal finance app built on a double-entry ledger (Go engine + Spring API + SvelteKit UI). Currency is Philippine Peso (PHP, `₱`, en-PH formatting).
Users track bank accounts, e-wallets, credit cards, and loans, record transactions as balanced debit/credit entries, and watch net worth.

Key pages: Dashboard (`/`), Accounts (`/accounts`, `/accounts/[id]`), Transactions (`/transactions`, `/transactions/[id]`, `/transactions/new`), Credit Cards (`/credit-cards`, planned).

JTBD
- "At a glance, how much do I have, how much do I owe, and what's my net worth?"
- "What moved recently, and in which account?"
- "Record a transaction quickly and correctly."

## Visual direction: refine the current look (do not replace it)
Dark-only UI. Calm, trustworthy, data-first. Dense but breathable. No marketing flourishes.

### Color (Tailwind v4 default palette)
- Page background: `slate-950` (#020617) with a subtle top radial glow `#172554 → #0b0f19` at 35%
- Surface / cards: `slate-900` (#0f172a), optionally `slate-900/80` with backdrop blur
- Raised / hover: `slate-800` (#1e293b)
- Borders: `slate-800`; hover border `slate-700`
- Text: primary `white`/`slate-50`, secondary `slate-400`, muted `slate-500`
- Accent (brand, primary buttons, links, focus): blue — `blue-600` buttons, hover `blue-500`, links `blue-400`
- Positive / inflow / assets: `emerald-400` text on `emerald-500/10` tint
- Negative / outflow / liabilities: `red-400` text on `red-500/10` tint
- Badges must use dark tints (`<color>-500/10` bg + `<color>-400` text), never light `-100` pills
- No other hues except for account-type badges: blue (credit card), purple (receivable), amber (payable), orange (expense)

### Typography
- Font: Inter (fallback system-ui). Only one family.
- Page title 28–30px bold, tight tracking. Section title 18px semibold. Body 14px. Labels 12px uppercase tracking-wide slate-400 for table headers.
- Money uses `tabular-nums`, bold, large for KPIs (30px).

### Shape, spacing, elevation
- Radius: `rounded-xl` buttons/inputs/icon tiles, `rounded-2xl` cards, `rounded-full` avatars and badges
- Card padding 20–24px; grid gaps 16px; page side padding 16/24/32px (sm/md/lg); container `max-w-7xl`
- Shadows: minimal; `shadow-xl` only on KPI cards; brand mark `shadow-lg shadow-blue-600/20`

### Icons
- Use lucide icons (lucide-svelte is installed): LayoutDashboard, Wallet, ArrowLeftRight, CreditCard, Plus, ArrowDownLeft, ArrowUpRight, TrendingUp, TrendingDown, Landmark, ChevronRight. No emoji, no unicode glyph icons.

### Components
- Navbar: h-16, brand mark (blue-600 rounded-xl square with "F") + "FinanKal", links Dashboard / Accounts / Transactions / Credit Cards, primary "New Transaction" button, user avatar.
- KPI card: label, big money value, short description, icon tile, optional delta/sparkline.
- List rows: icon circle tinted by direction, description + account • date, amount right-aligned with +/- sign and color.
- Buttons: primary blue-600 rounded-xl; secondary slate-800 with slate-700 border.

### Motion
- `transition` 150–200ms on hover color/border only. No bouncy or decorative animation.
