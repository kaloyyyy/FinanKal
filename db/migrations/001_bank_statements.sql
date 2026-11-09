-- ADR-0001: bank statements as stored records.
-- Run once on an existing database:  psql "$DATABASE_URL" -f db/migrations/001_bank_statements.sql
-- (Fresh Docker volumes get the same tables from db/schema.sql.)

BEGIN;

CREATE TABLE IF NOT EXISTS bank_statements
(
    id              uuid                     DEFAULT gen_random_uuid() NOT NULL PRIMARY KEY,
    account_id      uuid                                               NOT NULL REFERENCES accounts ON DELETE CASCADE,
    bank            varchar(20)                                        NOT NULL,
    period_start    date                                               NOT NULL,
    period_end      date                                               NOT NULL,
    opening_balance numeric(19, 4),
    closing_balance numeric(19, 4),
    file_sha256     char(64)                                           NOT NULL,
    imported_at     timestamp with time zone DEFAULT now()             NOT NULL,
    CONSTRAINT bank_statements_period_check CHECK (period_end >= period_start),
    CONSTRAINT bank_statements_file_unique UNIQUE (account_id, file_sha256),
    CONSTRAINT bank_statements_period_unique UNIQUE (account_id, period_start, period_end)
);

CREATE INDEX IF NOT EXISTS idx_bank_statements_account ON bank_statements (account_id, period_start);

ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS statement_id uuid REFERENCES bank_statements ON DELETE CASCADE,
    ADD COLUMN IF NOT EXISTS opening_balance boolean DEFAULT false NOT NULL;

CREATE INDEX IF NOT EXISTS idx_transactions_statement_id ON transactions (statement_id);

-- Corrections made by the previous importer were plain transactions named "Opening balance".
-- They are superseded by the single flagged opening balance; clear transactions and re-import statements.

COMMIT;
