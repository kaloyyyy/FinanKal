# Paikot Forecast Implementation Plan

> **For agentic workers:** Implement natively in the current workspace. Do not create/switch branches or worktrees, commit, or run tests/build/lint/previews.

**Goal:** Forecast linked repayments and card-funded installments into next month's bill while tracking available credit and principal-only limit replenishment.

**Architecture:** Extend the existing account/payable data path to expose card credit limits, statement billing days, and installment principal metadata where required. Add a dashboard Paikot card beside next month's dues, with browser-persisted forecast inputs and calculations that never create transactions.

**Tech Stack:** Go finance engine, Protocol Buffers, Spring API, SvelteKit/TypeScript, existing browser storage.

**Spec:** `docs/superpowers/specs/2026-11-11-paikot-forecast-design.md`

## Global Constraints

- Forecast only; no transaction creation or backend writes from Paikot.
- Full purchase price consumes available credit; only principal paid restores it.
- 0% plan: full installment payment is principal. Otherwise collect principal, or warn when skipped and assume full payment.
- Only forecast charges after the card statement close into the next billing month.
- PayDirect: show its supplied caps and fee estimate; flag loan/debt repayments as ineligible and use the explicit card-charge date for cycle placement.
- No new dependencies; current branch only; no tests/build/lint/previews or commits.

## Review Focus

- Existing balances must not be double-counted with installments already reflected in card utilization.
- Installment interest must not silently restore credit when a principal split is known.
- Dates on/before statement close must not be projected into the following statement.
- Skipped principal input must display the estimate warning.
- Invalid, missing, or negative limits/amounts must not produce misleading available credit.

## Files and responsibilities

- `proto/finance.proto`, generated `engine-go/finance/finance.pb.go`: carry card limit, statement billing day, and weekend shift through account data.
- `engine-go/cmd/server/server.go`: attach existing card terms to account responses.
- `api-spring` account DTO mapping and `finankal-ui/src/lib/types/account.ts`, server API: expose typed card forecast inputs.
- Existing `GET /api/users/{id}/payables` provides unpaid loan/card dues and statements; no payable schema change is needed.
- `finankal-ui/src/routes/+page.server.ts`, `src/routes/+page.svelte`, and new `src/lib/components/dashboard/PaikotForecast.svelte`: load the data and place the planner beside next month's dues.
- `src/lib/utils/paikotForecast.ts`: calculate utilization, payment posting/replenishment, and next-statement inclusion.

### Task 1: Expose card terms and installment details

- Extend account response fields with credit limit, billing day, and weekend statement shift, populated from existing card rows.
- Reuse existing payable records as forecast targets; keep their API unchanged.
- Map card terms through Spring account DTOs and UI account types.

### Task 2: Add forecast calculation

- Calculate baseline available credit from card limit and current balance.
- Apply scenario steps in date order: card-funded payment plus fee consumes limit; repayment posts back to target card; card installment payment restores only principal.
- Determine the next statement month using each card billing day and include only charges after that close.
- Enforce the forecast's PayDirect per-transaction, monthly count, recipient count, and monthly amount caps; flag debt repayment as ineligible.
- Keep unknown-principal fallback explicitly flagged.

### Task 3: Add dashboard Paikot card

- Put the card beside next month's bill in the existing dashboard forecast section.
- Allow selecting an existing payable or entering an external due, selecting a paying card, entering planned date/amount, and modeling repayment followed by reuse.
- Ask whether a plan is 0%; otherwise request monthly principal. If skipped, assume full payment as principal and display the warning.
- For PayDirect, collect fee percentage, recipient, scheduled payee date, and card-charge date; show the configured cap and disallow interpreting debt payables as PayDirect-eligible.
- Persist scenario inputs in browser storage only; do not call transaction endpoints.

### Completion

- Review changed files and diff only. Do not run tests, builds, lint, or previews per repository instructions.
- Do not commit.
