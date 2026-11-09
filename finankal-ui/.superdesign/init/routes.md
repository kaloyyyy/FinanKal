# Routes (SvelteKit file-based; all use src/routes/+layout.svelte)
| Path | File | Summary |
|---|---|---|
| `/` | `src/routes/+page.svelte` (+page.server.ts loads summary + 5 recent tx) | Dashboard: 3 SummaryCards (Liquid Assets, Total Liabilities, Net Worth), "Your Accounts" CTA card, RecentTransactions list |
| `/accounts` | `src/routes/accounts/+page.svelte` | Header + New Account button, 3 stat cards (total/asset/liability counts), accounts grouped by type in AccountCard grid |
| `/accounts/[id]` | `src/routes/accounts/[id]/+page.svelte` | AccountCard + AccountBalance + LedgerTable |
| `/transactions` | `src/routes/transactions/+page.svelte` | Transactions list with filters |
| `/transactions/[id]` | `src/routes/transactions/[id]/+page.svelte` | Transaction detail with entries |
| `/transactions/new` | `src/routes/transactions/new/+page.svelte` | Double-entry form: TransactionForm, TransactionEntriesTable, TransactionSummary, TransactionActions |
| `/credit-cards` | (linked in nav, no route yet) | — |
