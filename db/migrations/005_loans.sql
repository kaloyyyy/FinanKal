-- Loans whose interest is charged on what's still owed (Maya Personal Loan): paying ahead lowers the
-- principal and every later interest charge, so interest is booked per due date, not all up front.
BEGIN;

CREATE TABLE IF NOT EXISTS loans (
    account_id   uuid PRIMARY KEY REFERENCES accounts (id) ON DELETE CASCADE,
    booked_on    date           NOT NULL,
    daily_rate   numeric(10, 8) NOT NULL CHECK (daily_rate >= 0),
    -- What was owed at the end of principal_on: the loan amount on booking day, or a figure from the
    -- lender's app when the ledger can't rebuild it (penalties, the lender's payment order).
    principal_on date           NOT NULL,
    principal    numeric(19, 4) NOT NULL,
    file_sha256  char(64)       NOT NULL UNIQUE
);
-- Installments due on or before this date are paid, as the lender's app reports (its dues aren't the flat table).
ALTER TABLE loans ADD COLUMN IF NOT EXISTS paid_through date;

-- Transactions the engine books for a loan (interest on due dates and payment days, the catch-up to principal_on).
-- They're worked out again from the ledger on every read, so they're deleted and rebooked freely.
CREATE TABLE IF NOT EXISTS loan_charges (
    transaction_id uuid PRIMARY KEY REFERENCES transactions (id) ON DELETE CASCADE,
    account_id     uuid NOT NULL REFERENCES accounts (id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS loan_charges_account_id ON loan_charges (account_id);

COMMIT;
