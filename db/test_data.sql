-- =========================
-- CLEAR TABLES
-- =========================
TRUNCATE TABLE
    credit_card_entries,
    credit_card_statements,
    credit_cards,
    entries,
    transactions,
    accounts,
    users
    RESTART IDENTITY CASCADE;

-- =========================
-- USERS
-- =========================
INSERT INTO users (id, name, username)
VALUES
    (
        'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
        'John Doe',
        'johndoe'
    ),
    (
        'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
        'Jane Smith',
        'janesmith'
    );

-- =========================
-- ACCOUNTS
-- =========================
INSERT INTO accounts (id, user_id, name, type)
VALUES
    (
        '11111111-1111-1111-1111-111111111111',
        'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
        'Checking Account',
        'ASSET'
    ),
    (
        '22222222-2222-2222-2222-222222222222',
        'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
        'Savings Account',
        'ASSET'
    ),
    (
        '33333333-3333-3333-3333-333333333333',
        'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
        'Business Account',
        'ASSET'
    ),
    (
        '44444444-4444-4444-4444-444444444444',
        'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
        'BPI Visa',
        'CREDIT_CARD'
    ),
    (
        '55555555-5555-5555-5555-555555555555',
        'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
        'Food Expense',
        'EXPENSE'
    );

-- =========================
-- CREDIT CARD
-- =========================
INSERT INTO credit_cards
(
    id,
    account_id,
    credit_limit,
    billing_day,
    payment_due_days
)
VALUES
    (
        '1a1a1a1a-1a1a-1a1a-1a1a-1a1a1a1a1a1a',
        '44444444-4444-4444-4444-444444444444',
        100000.00,
        15,
        21
    );

-- =========================
-- NORMAL TRANSACTIONS
-- =========================
INSERT INTO transactions
(
    id,
    user_id,
    description
)
VALUES
    (
        '66666666-6666-6666-6666-666666666666',
        'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
        'Initial Deposit'
    ),
    (
        '77777777-7777-7777-7777-777777777777',
        'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
        'Transfer'
    );

-- =========================
-- NORMAL ENTRIES
-- =========================
INSERT INTO entries
(
    id,
    transaction_id,
    account_id,
    amount,
    type
)
VALUES
    (
        'eeeeeeee-1111-1111-1111-111111111111',
        '66666666-6666-6666-6666-666666666666',
        '11111111-1111-1111-1111-111111111111',
        1000,
        'DEBIT'
    ),
    (
        'eeeeeeee-2222-2222-2222-222222222222',
        '66666666-6666-6666-6666-666666666666',
        '22222222-2222-2222-2222-222222222222',
        1000,
        'CREDIT'
    ),
    (
        'eeeeeeee-3333-3333-3333-333333333333',
        '77777777-7777-7777-7777-777777777777',
        '11111111-1111-1111-1111-111111111111',
        250,
        'DEBIT'
    ),
    (
        'eeeeeeee-4444-4444-4444-444444444444',
        '77777777-7777-7777-7777-777777777777',
        '22222222-2222-2222-2222-222222222222',
        250,
        'CREDIT'
    );

-- =========================
-- CREDIT CARD PURCHASE
-- =========================
INSERT INTO transactions
(
    id,
    user_id,
    description
)
VALUES
    (
        '88888888-8888-8888-8888-888888888888',
        'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
        'Jollibee'
    );

INSERT INTO entries
(
    id,
    transaction_id,
    account_id,
    amount,
    type
)
VALUES
    (
        'eeeeeeee-5555-5555-5555-555555555555',
        '88888888-8888-8888-8888-888888888888',
        '44444444-4444-4444-4444-444444444444',
        350,
        'CREDIT'
    );

-- =========================
-- CREDIT CARD STATEMENT
-- =========================
INSERT INTO credit_card_statements
(
    id,
    credit_card_id,
    start_date,
    end_date,
    statement_date,
    due_date,
    total_amount,
    status
)
VALUES
    (
        '99999999-9999-9999-9999-999999999999',
        '1a1a1a1a-1a1a-1a1a-1a1a-1a1a1a1a1a1a',
        '2026-08-01',
        '2026-08-15',
        '2026-08-15',
        '2026-09-05',
        350,
        'OPEN'
    );

-- =========================
-- MAP ENTRY TO STATEMENT
-- =========================
INSERT INTO credit_card_entries
(
    statement_id,
    entry_id
)
VALUES
    (
        '99999999-9999-9999-9999-999999999999',
        'eeeeeeee-5555-5555-5555-555555555555'
    );