-- A fourth due rule, MariBank's credit card:
--   DAYS_SKIP_WEEKEND payment_due_days calendar days after the statement date, or the Monday after when that's a weekend
BEGIN;

ALTER TABLE credit_cards DROP CONSTRAINT IF EXISTS credit_cards_due_rule_check;
ALTER TABLE credit_cards ADD CONSTRAINT credit_cards_due_rule_check
    CHECK (due_rule IN ('DAYS', 'DAY_OF_MONTH', 'BUSINESS_DAYS', 'DAYS_SKIP_WEEKEND'));

COMMIT;
