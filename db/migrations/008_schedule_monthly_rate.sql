-- A card installment plan's monthly interest rate (0.0178805 = 1.78805%). The card's limit is held by the
-- principal still unbilled, not by the interest in each installment; 0 holds the whole amount.
BEGIN;

ALTER TABLE scheduled_transactions
    ADD COLUMN IF NOT EXISTS monthly_rate numeric(10, 8) NOT NULL DEFAULT 0 CHECK (monthly_rate >= 0);

COMMIT;
