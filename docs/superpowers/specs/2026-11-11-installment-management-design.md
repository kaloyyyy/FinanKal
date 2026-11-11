# Installment Management Design

## Intent

Let users record scheduled loan and credit card installment payments, update plans, and cancel plans while keeping completed payments and ledger history intact.

## Scope

- Allow the existing manual payment action for the next installment of a loan or credit card, using its saved bank/wallet source.
- Complete card schedule occurrences when payment is recorded; leave loan settlement to the existing ledger-based loan logic.
- Edit schedule name, source, and pending due dates/amounts. Keep completed occurrences and their transactions unchanged.
- Edit loan account name and imported terms; new terms affect recalculated future interest while existing disbursement entries remain unchanged.
- Delete a loan by canceling its schedule and preserving its account and ledger history.
- Delete an installment plan by deactivating it and canceling pending occurrences.

## Non-goals

No automatic transactions, hard deletion of loan accounts, removal of ledger history, or editing completed installment payments.

## Approaches

Extend the existing schedule and loan APIs and UI, or introduce a separate installment subsystem. The existing schedule already stores both loan and card installments, so extending it keeps one source of truth and avoids duplicating payment logic.

## Constraints and calls

Use the current Go engine, protobuf bridge, Spring API, and Svelte account page patterns. Mutations are scoped to the authenticated user's records. Completed schedule rows and ledger transactions are immutable through these controls. Do not run tests, builds, lint, or previews.
