# ADR-0001: Bank statements as stored records, imported in one engine transaction

**Status:** Accepted
**Date:** 2026-11-09
**Deciders:** kaloyyyy

## Context

FinanKal imports PDF bank statements (MariBank first; BDO, BPI, GCash, Maya and GoTyme planned) so spending
doesn't have to be typed in. A MariBank statement gives a period (`01 SEP 2026 to 30 SEP 2026`), a starting
and ending balance, and every row with a date, counterparty, type and an Outgoing or Incoming amount.
Consecutive statements chain: September ends at ₱24,849.21 and October starts at ₱24,849.21.

The first implementation lived entirely in Spring and treated a statement as nothing more than rows:

- **Balances drifted.** Only the rows were recorded, so a cleared account ended at
  `incoming − outgoing` (−₱23k) instead of the statement's ending balance (₱1,496.63).
- **Order mattered.** The fix compared the opening balance with the *current* balance, so importing
  September after October left the account at September's closing balance.
- **The workaround was fragile.** Per-month "Opening balance" corrections, found again by matching their
  description, were deleted and recreated by Spring over many separate gRPC calls.
- **Nothing stopped duplicates.** Importing a file twice recorded every row twice.
- **Imports weren't atomic.** A failure halfway left part of a statement in the ledger.
- **Errors went unnoticed.** Nothing checked the result against the statement's own totals, so a parser
  miss (e.g. MariBank's day-less interest row) showed up only as a wrong balance later.

## Decision

1. **Store each statement** in a `bank_statements` table: account, bank, period, opening and closing
   balance, and the file's SHA-256. Rows imported from it carry `transactions.statement_id`.
   The same file, or a second statement for the same period, is refused per account.
2. **One opening balance per account.** A single transaction flagged `transactions.opening_balance` is
   dated at the start of the account's *earliest* statement, for
   `that statement's opening balance − entries dated before it`. It is recalculated on every import,
   so import order no longer matters.
3. **Report, don't paper over.** After importing, every statement of the account is reconciled: the ledger
   balance at its period start and end must equal its printed opening and closing balance. A mismatch
   (missing month, parser miss, manual entry) comes back as a warning; the ledger is not adjusted to hide it.
4. **Import in one engine call, in one database transaction**: `ImportBankStatement` locks the account,
   records the statement and its rows, finds or creates the counter accounts (expense, income, equity), and
   recalculates the opening balance, or does nothing at all.
5. **Spring only parses.** It extracts the PDF, hashes it, checks the account belongs to the user, and sends
   the rows. The per-month correction logic, `DeleteTransaction` and `CreateAccount` are removed.

## Options considered

### A. Keep Spring-side corrections (previous implementation)
| Dimension | Assessment |
|-----------|------------|
| Complexity | Medium, and spread across services |
| Correctness | Order-independent only by deleting/recreating corrections; not atomic |
| Duplicates | Not detected |
| Change cost | None |

**Pros:** already written. **Cons:** every problem listed in Context remains except order; corrections are
recognised by description; dozens of gRPC calls per import.

### B. Statements as records, one opening balance, atomic engine import (chosen)
| Dimension | Assessment |
|-----------|------------|
| Complexity | Low–medium, in one place (engine) |
| Correctness | Atomic; order-independent by construction; reconciled against the statement |
| Duplicates | Refused by unique constraints |
| Change cost | One table, two columns, one RPC; removes two RPCs and the Spring correction code |

**Pros:** the ledger can always be checked against the bank; undo-by-statement becomes possible.
**Cons:** schema migration; statement logic now lives in Go.

### C. Opening balance as an account attribute (no transaction)
Store `accounts.opening_balance` and add it when computing balances.

**Pros:** no synthetic transaction. **Cons:** every balance query in the engine (balance, net worth,
totals, ledger) would need to include it; double-entry no longer balances against equity. Rejected.

## Trade-off analysis

B costs a migration but turns "is the ledger right?" from a guess into a check, which is the point of
importing bank statements. C is cleaner in theory but touches every balance query for no user-visible gain.
A is cheapest today and the most expensive to trust.

## Consequences

- **Easier:** importing months in any order; spotting a missing month; trusting balances; deleting one
  statement later (its transactions cascade).
- **Harder:** statements that don't print an opening balance (some e-wallet exports) still import their
  rows, but can't anchor the balance or be reconciled.
- **Revisit:** credit card statements (balance means *owed*, not *held*) need their own rules; only
  bank / wallet (`ASSET`) accounts are accepted for now. Counter accounts are the user's first
  expense / income / equity account until categories exist.

## Action items

1. [x] `bank_statements` table, `transactions.statement_id` and `transactions.opening_balance`
   (`db/migrations/001_bank_statements.sql`, `db/schema.sql`).
2. [x] `ImportBankStatement` RPC in the engine, with reconciliation warnings.
3. [x] Clearing transactions also clears statements, so they can be re-imported.
4. [x] Spring sends parsed rows, period, balances and file hash; import page shows warnings.
5. [ ] Delete a single statement from the UI.
6. [ ] Credit card statement import.
