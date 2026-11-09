# Page dependency trees

## / (Dashboard)
Entry: src/routes/+page.svelte
- src/routes/+layout.svelte
  - src/routes/layout.css
  - src/lib/components/layout/Navbar.svelte
- src/lib/components/dashboard/SummaryCards.svelte
  - src/lib/components/dashboard/SummaryCard.svelte
- src/lib/components/dashboard/RecentTransactions.svelte
  - src/lib/utils/currency.ts
  - src/lib/server/api/transactions.ts (type RecentTransaction)

## /accounts
Entry: src/routes/accounts/+page.svelte
- src/lib/components/accounts/AccountCard.svelte
  - src/lib/components/accounts/AccountTypeBadge.svelte
  - src/lib/utils/currency.ts

## /accounts/[id]
Entry: src/routes/accounts/[id]/+page.svelte
- src/lib/components/accounts/AccountCard.svelte
  - src/lib/components/accounts/AccountTypeBadge.svelte
- src/lib/components/accounts/AccountBalance.svelte
- src/lib/components/accounts/LedgerTable.svelte
  - src/lib/types/ledger.ts

## /transactions
Entry: src/routes/transactions/+page.svelte (self-contained)

## /transactions/[id]
Entry: src/routes/transactions/[id]/+page.svelte (self-contained)

## /transactions/new
Entry: src/routes/transactions/new/+page.svelte
- src/lib/components/transactions/TransactionForm.svelte
- src/lib/components/transactions/TransactionEntriesTable.svelte
  - src/lib/components/transactions/TransactionEntryRow.svelte
- src/lib/components/transactions/TransactionSummary.svelte
- src/lib/components/transactions/TransactionActions.svelte
