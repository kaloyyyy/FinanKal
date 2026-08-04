-- Database schema for FinanKal

-- =========================
-- USERS
-- =========================
CREATE TABLE IF NOT EXISTS users
(
    id         UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    name       VARCHAR(255)        NOT NULL,
    username   VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- =========================
-- ACCOUNTS
-- =========================
CREATE TABLE IF NOT EXISTS accounts
(
    id         UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    user_id    UUID         NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name       VARCHAR(255) NOT NULL,
    type       VARCHAR(50)  NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Ensure user_id exists (for older schemas)
ALTER TABLE accounts
    ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users (id) ON DELETE CASCADE;

-- Optional: enforce account types
ALTER TABLE accounts
    DROP CONSTRAINT IF EXISTS chk_account_type;

ALTER TABLE accounts
    ADD CONSTRAINT chk_account_type
        CHECK (type IN ('ASSET', 'LIABILITY', 'EXPENSE', 'INCOME', 'EQUITY', 'CREDIT_CARD'));

-- =========================
-- TRANSACTIONS
-- =========================
CREATE TABLE IF NOT EXISTS transactions
(
    id          UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    description TEXT,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Ensure user_id exists (for older schemas)
ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users (id) ON DELETE CASCADE;

-- =========================
-- ENTRIES (DOUBLE-ENTRY LEDGER)
-- =========================
CREATE TABLE IF NOT EXISTS entries
(
    id             UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    transaction_id UUID           NOT NULL REFERENCES transactions (id) ON DELETE CASCADE,
    account_id     UUID           NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    amount         DECIMAL(19, 4) NOT NULL,
    type           VARCHAR(10)    NOT NULL CHECK (type IN ('DEBIT', 'CREDIT')),
    created_at     TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- =========================
-- CREDIT CARDS
-- =========================
CREATE TABLE IF NOT EXISTS credit_cards
(
    id           UUID PRIMARY KEY         DEFAULT gen_random_uuid(),

    account_id   UUID           NOT NULL UNIQUE
        REFERENCES accounts (id) ON DELETE CASCADE,

    credit_limit NUMERIC(15, 2) NOT NULL,

    billing_day  INT            NOT NULL CHECK (billing_day BETWEEN 1 AND 28),
    due_day      INT            NOT NULL CHECK (due_day BETWEEN 1 AND 28),

    cutoff_time  TIME                     DEFAULT '23:59:59',

    created_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- =========================
-- CREDIT CARD STATEMENTS
-- =========================
CREATE TABLE IF NOT EXISTS credit_card_statements
(
    id             UUID PRIMARY KEY         DEFAULT gen_random_uuid(),

    credit_card_id UUID        NOT NULL
        REFERENCES credit_cards (id) ON DELETE CASCADE,

    start_date     DATE        NOT NULL,
    end_date       DATE        NOT NULL,

    statement_date DATE        NOT NULL,
    due_date       DATE        NOT NULL,

    total_amount   NUMERIC(15, 2)           DEFAULT 0,

    status         VARCHAR(20) NOT NULL     DEFAULT 'OPEN'
        CHECK (status IN ('OPEN', 'CLOSED', 'PAID')),

    created_at     TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at     TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT unique_statement_cycle
        UNIQUE (credit_card_id, start_date, end_date)
);

-- =========================
-- STATEMENT ↔ ENTRIES MAPPING (CLEAN DESIGN)
-- =========================
CREATE TABLE IF NOT EXISTS credit_card_entries
(
    id           UUID PRIMARY KEY         DEFAULT gen_random_uuid(),

    statement_id UUID NOT NULL
        REFERENCES credit_card_statements (id) ON DELETE CASCADE,

    entry_id     UUID NOT NULL
        REFERENCES entries (id) ON DELETE CASCADE,

    created_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT unique_statement_entry
        UNIQUE (statement_id, entry_id)
);

-- =========================
-- EXTERNAL ACCOUNTS (BANK API READY)
-- =========================
CREATE TABLE IF NOT EXISTS external_accounts
(
    id             UUID PRIMARY KEY         DEFAULT gen_random_uuid(),

    account_id     UUID        NOT NULL
        REFERENCES accounts (id) ON DELETE CASCADE,

    provider       VARCHAR(50) NOT NULL,
    external_id    VARCHAR(255),

    access_token   TEXT,
    refresh_token  TEXT,

    last_synced_at TIMESTAMP WITH TIME ZONE,

    created_at     TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at     TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- =========================
-- INDEXES (PERFORMANCE)
-- =========================

-- Accounts & users
CREATE INDEX IF NOT EXISTS idx_accounts_user_id ON accounts (user_id);

-- Transactions
CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions (user_id);

-- Entries
CREATE INDEX IF NOT EXISTS idx_entries_account_id ON entries (account_id);
CREATE INDEX IF NOT EXISTS idx_entries_transaction_id ON entries (transaction_id);

-- Statements
CREATE INDEX IF NOT EXISTS idx_cc_statement_card_dates
    ON credit_card_statements (credit_card_id, start_date, end_date);

CREATE INDEX IF NOT EXISTS idx_cc_statement_status
    ON credit_card_statements (status);

-- Statement ↔ Entries mapping
CREATE INDEX IF NOT EXISTS idx_stmt_entries_statement
    ON credit_card_entries (statement_id);

CREATE INDEX IF NOT EXISTS idx_stmt_entries_entry
    ON credit_card_entries (entry_id);