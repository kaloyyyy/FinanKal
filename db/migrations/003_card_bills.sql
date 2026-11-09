-- Card bills the engine tracks: one per billing cycle, with a due date either printed on the statement
-- or worked out from the card's rule. Whether a bill is paid is worked out from the ledger when read.
BEGIN;

-- How a card's due date follows its statement date:
--   DAYS          payment_due_days calendar days after the statement date
--   DAY_OF_MONTH  the next payment_due_days-th of a month after the statement date (e.g. the 5th)
--   BUSINESS_DAYS payment_due_days weekdays after the statement date
ALTER TABLE credit_cards ADD COLUMN IF NOT EXISTS due_rule varchar(20) NOT NULL DEFAULT 'DAYS';
ALTER TABLE credit_cards DROP CONSTRAINT IF EXISTS credit_cards_due_rule_check;
ALTER TABLE credit_cards ADD CONSTRAINT credit_cards_due_rule_check CHECK (due_rule IN ('DAYS', 'DAY_OF_MONTH', 'BUSINESS_DAYS'));

-- A bill without a printed due date takes the card's rule.
ALTER TABLE credit_card_statements ALTER COLUMN due_date DROP NOT NULL;

-- One bill per card per statement date, so re-importing a statement updates it instead of adding another.
CREATE UNIQUE INDEX IF NOT EXISTS credit_card_statements_card_date_unique
    ON credit_card_statements (credit_card_id, statement_date);

COMMIT;
