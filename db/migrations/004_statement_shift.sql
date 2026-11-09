-- When a card's billing day falls on a weekend, its statement date moves:
--   NONE                  stays on the billing day (app-based BNPL like Atome)
--   NEXT_BUSINESS_DAY     the following weekday (most banks)
--   NEAREST_BUSINESS_DAY  Saturday back to Friday, Sunday forward to Monday
-- Bills the engine works out before the statement arrives are saved with status ESTIMATED and
-- replaced when the statement is imported.
BEGIN;

ALTER TABLE credit_cards ADD COLUMN IF NOT EXISTS statement_shift varchar(25) NOT NULL DEFAULT 'NEXT_BUSINESS_DAY';
ALTER TABLE credit_cards DROP CONSTRAINT IF EXISTS credit_cards_statement_shift_check;
ALTER TABLE credit_cards ADD CONSTRAINT credit_cards_statement_shift_check
    CHECK (statement_shift IN ('NONE', 'NEXT_BUSINESS_DAY', 'NEAREST_BUSINESS_DAY'));

COMMIT;
