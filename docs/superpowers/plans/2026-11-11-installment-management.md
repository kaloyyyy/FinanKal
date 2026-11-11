# Installment Management Implementation Plan

> **For agentic workers:** Native implementation in this session. Follow the approved design and the repository's existing Go, protobuf, Spring, and Svelte patterns.

**Goal:** Pay, edit, and cancel loan and credit card installment plans while retaining completed payments and ledger history.

**Architecture:** Extend schedule and loan operations in the Go engine, expose them through protobuf/Spring, then connect the existing account page and deadlines component. Payment creation remains manual and uses server-loaded schedule details.

**Tech Stack:** Go, PostgreSQL, protobuf/gRPC, Spring Boot, SvelteKit/TypeScript.

**Spec:** `docs/superpowers/specs/2026-11-11-installment-management-design.md`

## Global Constraints

- Preserve completed occurrences and all ledger transactions.
- Scope every update or cancellation by user and record ID.
- Do not run tests, builds, lint, or previews.

## Review Focus

- A caller cannot edit or cancel another user's schedule or loan.
- Editing retains completed occurrences while replacing only pending rows.
- Card payment closes exactly its occurrence; loan payment continues using existing settlement logic.
- Deleting a loan keeps its account, charges, and ledger history.
- A missing or invalid saved funding account cannot create an unbalanced payment.

---

### Task 1: Schedule management and payment identity

**Files:** `engine-go/internal/schedule/repository.go`, `engine-go/internal/schedule/service.go`, `engine-go/finance/finance.proto`, generated Go protobuf sources, Go API server and Spring DTO/service/controller.

- [ ] Expose schedule and occurrence IDs on payable records.
- [ ] Add user-scoped update and cancel operations. Preserve completed occurrences; replace only pending installments during update and deactivate/cancel pending rows on delete.
- [ ] Add an atomic scheduled card payment operation that validates the pending occurrence and saved accounts, creates balanced ledger entries, and completes that occurrence in one database transaction.
- [ ] Keep loan schedule payments routed through the existing loan settlement behavior.
- [ ] Map validation and ownership failures to existing API error conventions.

### Task 2: Loan detail editing and cancellation

**Files:** `engine-go/internal/loan/service.go`, protobuf, Go API server, Spring API DTO/service/controller.

- [ ] Add user-scoped loan detail reads needed by the editor and updates for name, booked date, principal date/amount, and daily rate.
- [ ] Leave original disbursement transactions unchanged; update only the loan terms used for future accrued charge recalculation.
- [ ] Cancel a loan's associated schedule and its pending occurrences while retaining the account and ledger history.

### Task 3: Account page controls

**Files:** `finankal-ui/src/lib/utils/payables.ts`, `finankal-ui/src/lib/server/api/payables.ts`, `finankal-ui/src/routes/accounts/[id]/+page.server.ts`, account page Svelte files.

- [ ] Enable the existing manual payment checkbox for the next saved-source loan or card installment; use the server's payable identity and amount when posting.
- [ ] Add schedule edit/delete controls and loan detail edit/delete controls using the new API operations.
- [ ] Refresh account deadlines after mutations and surface API errors accessibly.
- [ ] Do not run test, build, lint, or preview commands, per workspace instructions.
