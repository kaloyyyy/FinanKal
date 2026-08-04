-- Test data for FinanKal

-- =========================
-- USERS
-- =========================
INSERT INTO users (id, name, username, created_at)
VALUES ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'John Doe', 'johndoe', NOW()),
       ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Jane Smith', 'janesmith', NOW());

-- =========================
-- ACCOUNTS
-- =========================
INSERT INTO accounts (id, user_id, name, type, created_at)
VALUES ('11111111-1111-1111-1111-111111111111', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Checking Account', 'ASSET',
        NOW()),
       ('22222222-2222-2222-2222-222222222222', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Savings Account', 'ASSET',
        NOW()),
       ('33333333-3333-3333-3333-333333333333', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Business Account', 'ASSET',
        NOW()),

-- Credit Card account
       ('44444444-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'BPI Credit Card',
        'CREDIT_CARD', NOW()),

-- Expense account
       ('55555555-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Food Expense', 'EXPENSE',
        NOW());

-- =========================
-- CREDIT CARD CONFIG
-- =========================
INSERT INTO credit_cards (id,
                          account_id,
                          credit_limit,
                          billing_day,
                          due_day,
                          created_at)
VALUES ('cccccccc-cccc-cccc-cccc-cccccccccccc',
        '44444444-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
        100000.00,
        15,
        5,
        NOW());

-- =========================
-- TRANSACTIONS (NORMAL)
-- =========================
INSERT INTO transactions (id, user_id, description, created_at)
VALUES ('44444444-4444-4444-4444-444444444444', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Initial deposit', NOW()),
       ('55555555-5555-5555-5555-555555555555', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Transfer between accounts',
        NOW());

-- =========================
-- ENTRIES (NORMAL)
-- =========================
INSERT INTO entries (transaction_id, account_id, amount, type, created_at)
VALUES ('44444444-4444-4444-4444-444444444444', '11111111-1111-1111-1111-111111111111', 500.00, 'DEBIT', NOW()),
       ('44444444-4444-4444-4444-444444444444', '22222222-2222-2222-2222-222222222222', 500.00, 'CREDIT', NOW()),
       ('55555555-5555-5555-5555-555555555555', '11111111-1111-1111-1111-111111111111', 100.00, 'DEBIT', NOW()),
       ('55555555-5555-5555-5555-555555555555', '22222222-2222-2222-2222-222222222222', 100.00, 'CREDIT', NOW());

-- =========================
-- CREDIT CARD STATEMENT
-- =========================
INSERT INTO credit_card_statements (id,
                                    credit_card_id,
                                    start_date,
                                    end_date,
                                    statement_date,
                                    due_date,
                                    total_amount,
                                    status,
                                    created_at)
VALUES ('dddddddd-dddd-dddd-dddd-dddddddddddd',
        'cccccccc-cccc-cccc-cccc-cccccccccccc',
        '2026-03-16',
        '2026-04-15',
        '2026-04-15',
        '2026-05-05',
        0,
        'OPEN',
        NOW());

-- =========================
-- CREDIT CARD TRANSACTION
-- =========================
INSERT INTO transactions (id, user_id, description, created_at)
VALUES ('66666666-6666-6666-6666-666666666666',
        'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
        'Jollibee order',
        '2026-04-10');

-- =========================
-- CREDIT CARD ENTRIES
-- =========================
-- Expense side
INSERT INTO entries (id, transaction_id, account_id, amount, type, created_at)
VALUES ('eeeeeeee-0000-0000-0000-000000000001',
        '66666666-6666-6666-6666-666666666666',
        '55555555-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
        250.00,
        'DEBIT',
        NOW());

-- Credit card liability side
INSERT INTO entries (id, transaction_id, account_id, amount, type, created_at)
VALUES ('eeeeeeee-0000-0000-0000-000000000002',
        '66666666-6666-6666-6666-666666666666',
        '44444444-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
        250.00,
        'CREDIT',
        NOW());

-- =========================
-- CREDIT CARD ENTRY MAPPING (NEW ARCHITECTURE)
-- =========================
INSERT INTO credit_card_entries (credit_card_id,
                                 statement_id,
                                 entry_id)
VALUES ('cccccccc-cccc-cccc-cccc-cccccccccccc',
        'dddddddd-dddd-dddd-dddd-dddddddddddd',
        'eeeeeeee-0000-0000-0000-000000000002');