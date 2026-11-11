# Loan installment payment checkbox

## Intent
Let the user pay an upcoming loan installment directly from the loan's Repayment list. Reuse the wallet/bank source saved with that installment schedule so the transaction reduces the matching source balance and loan balance together.

## Goals and constraints
- Show a payment checkbox/action beside the oldest unpaid installment for a loan.
- Checking it records a balanced transaction: debit the loan and credit its scheduled source wallet/bank, for the payable amount, dated today.
- Use the existing loan settlement calculation so the paid installment disappears from pending payables.
- Keep later installments unavailable until older installments are paid, matching the engine's oldest-first settlement behavior.
- Do not add an independent paid flag; the ledger transaction remains the source of truth.

## Approaches
- **Pay beside the installment (selected):** use its schedule-linked source account and the existing transaction endpoint. This matches the user's choice and removes source-account re-entry.
- **Tag a transaction on the general form:** requires selecting the loan and source manually and does not use the installment schedule's saved source.

## Implementation calls
- Include the schedule source account with loan payable data so the repayment row can use it.
- Expose the action only for loan installments with a saved source account; leave credit-card bill pay behavior unchanged.
- On submit, re-fetch the payable on the server and use its current amount and saved source account; do not trust values posted by the browser.
- On success refresh the loan account page so existing settlement logic recomputes remaining installments.
- Show an in-progress state and surface transaction failures without clearing the installment.

## Out of scope
Editing schedules, partial payments, paying out of order, and changing how loan interest or installment settlement is calculated.
