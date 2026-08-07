TRUNCATE TABLE
    credit_card_entries,
    credit_card_statements,
    credit_cards,
    entries,
    transactions,
    accounts,
    users
    RESTART IDENTITY CASCADE;


DO
$$
    DECLARE
        uid              UUID := 'bcc539f6-fc94-4a2e-9ce6-3c03c0560204';
        unionbank_id     UUID;
        gcash_id         UUID;
        maya_id          UUID;
        maribank_id      UUID;
        ub_cc_id         UUID;
        spay_id          UUID;
        atome_id         UUID;
        maribank_loan_id UUID;
        ub_clear_id      UUID;
        spay_clear_id    UUID;
        atome_clear_id   UUID;
        loan_clear_id    UUID;

    BEGIN


        -- USER
        INSERT INTO users(name, username, id)
        VALUES ('Kaloy', 'kal_6a',
                'bcc539f6-fc94-4a2e-9ce6-3c03c0560204');


        INSERT INTO accounts(user_id, name, type)
        VALUES ('bcc539f6-fc94-4a2e-9ce6-3c03c0560204',
                'Opening Balance Equity',
                'EQUITY');
        -- =========================
-- ASSETS
-- =========================

        INSERT INTO accounts(user_id, name, type)
        VALUES (uid, 'UnionBank Payroll', 'ASSET')
        RETURNING id INTO unionbank_id;


        INSERT INTO accounts(user_id, name, type)
        VALUES (uid, 'GCash', 'ASSET')
        RETURNING id INTO gcash_id;


        INSERT INTO accounts(user_id, name, type)
        VALUES (uid, 'Maya', 'ASSET')
        RETURNING id INTO maya_id;


        INSERT INTO accounts(user_id, name, type)
        VALUES (uid, 'MariBank', 'ASSET')
        RETURNING id INTO maribank_id;


        -- =========================
-- CLEARING ACCOUNTS
-- =========================

        INSERT INTO accounts(user_id, name, type)
        VALUES (uid, 'UnionBank Platinum CC Clearing', 'CLEARING')
        RETURNING id INTO ub_clear_id;


        INSERT INTO accounts(user_id, name, type)
        VALUES (uid, 'SPayLater Clearing', 'CLEARING')
        RETURNING id INTO spay_clear_id;


        INSERT INTO accounts(user_id, name, type)
        VALUES (uid, 'Atome Clearing', 'CLEARING')
        RETURNING id INTO atome_clear_id;


        INSERT INTO accounts(user_id, name, type)
        VALUES (uid, 'MariBank Loan Clearing', 'CLEARING')
        RETURNING id INTO loan_clear_id;


        -- =========================
-- CREDIT ACCOUNTS
-- =========================


        INSERT INTO accounts(user_id, name, type)
        VALUES (uid, 'UnionBank Platinum CC', 'CREDIT_CARD')
        RETURNING id INTO ub_cc_id;


        INSERT INTO accounts(user_id, name, type)
        VALUES (uid, 'SPayLater', 'CREDIT_CARD')
        RETURNING id INTO spay_id;


        INSERT INTO accounts(user_id, name, type)
        VALUES (uid, 'Atome', 'CREDIT_CARD')
        RETURNING id INTO atome_id;


        INSERT INTO accounts(user_id, name, type)
        VALUES (uid, 'MariBank Loan', 'LIABILITY')
        RETURNING id INTO maribank_loan_id;


        -- =========================
-- CREDIT CARD DETAILS
-- =========================


        INSERT INTO credit_cards(id,
                                 clearing_account_id,
                                 credit_limit,
                                 billing_day,
                                 payment_due_days)
        VALUES (ub_cc_id,
                ub_clear_id,
                35000,
                24,
                14);


        INSERT INTO credit_cards(id,
                                 clearing_account_id,
                                 credit_limit,
                                 billing_day,
                                 payment_due_days)
        VALUES (spay_id,
                spay_clear_id,
                17500,
                24,
                11);


        INSERT INTO credit_cards(id,
                                 clearing_account_id,
                                 credit_limit,
                                 billing_day,
                                 payment_due_days)
        VALUES (atome_id,
                atome_clear_id,
                35000,
                24,
                14);


        INSERT INTO credit_cards(id,
                                 clearing_account_id,
                                 credit_limit,
                                 billing_day,
                                 payment_due_days)
        VALUES (maribank_loan_id,
                loan_clear_id,
                45000,
                1,
                3);


    END
$$;

DO
$$

    DECLARE

        v_user_id      UUID := 'bcc539f6-fc94-4a2e-9ce6-3c03c0560204';
        equity_account UUID;
        maribank       UUID;
        maribank_loan  UUID;
        unionbank_cc   UUID;
        atome          UUID;
        spaylater      UUID;
        tx             UUID;

    BEGIN


        SELECT id
        INTO equity_account
        FROM accounts
        WHERE accounts.user_id = v_user_id
          AND accounts.name = 'Opening Balance Equity';


        SELECT id
        INTO maribank
        FROM accounts
        WHERE accounts.user_id = v_user_id
          AND accounts.name = 'MariBank';


        SELECT id
        INTO maribank_loan
        FROM accounts
        WHERE accounts.user_id = v_user_id
          AND accounts.name = 'MariBank Loan';


        SELECT id
        INTO unionbank_cc
        FROM accounts
        WHERE accounts.user_id = v_user_id
          AND accounts.name = 'UnionBank Platinum CC';


        SELECT id
        INTO atome
        FROM accounts
        WHERE accounts.user_id = v_user_id
          AND accounts.name = 'Atome';


        SELECT id
        INTO spaylater
        FROM accounts
        WHERE accounts.user_id = v_user_id
          AND accounts.name = 'SPayLater';

        -- =========================
-- MariBank Asset
-- 4,882.58
-- =========================

        INSERT INTO transactions(user_id,
                                 description)
        VALUES (v_user_id,
                'Opening Balance - MariBank')
        RETURNING id INTO tx;

        INSERT INTO entries(transaction_id,
                            account_id,
                            amount,
                            type)
        VALUES (tx,
                maribank,
                4882.58,
                'DEBIT'),
               (tx,
                equity_account,
                4882.58,
                'CREDIT');


        -- =========================
-- MariBank Loan
-- 3,628.33
-- =========================

        INSERT INTO transactions(user_id,
                                 description)
        VALUES (v_user_id,
                'Opening Balance - MariBank Loan')
        RETURNING id INTO tx;


        INSERT INTO entries(transaction_id,
                            account_id,
                            amount,
                            type)
        VALUES (tx,
                equity_account,
                3628.33,
                'DEBIT'),
               (tx,
                maribank_loan,
                3628.33,
                'CREDIT');


        -- =========================
-- UnionBank Platinum CC
-- 9,900
-- =========================

        INSERT INTO transactions(user_id,
                                 description)
        VALUES (v_user_id,
                'Opening Balance - UnionBank Platinum CC')
        RETURNING id INTO tx;


        INSERT INTO entries(transaction_id,
                            account_id,
                            amount,
                            type)
        VALUES (tx,
                equity_account,
                9900,
                'DEBIT'),
               (tx,
                unionbank_cc,
                9900,
                'CREDIT');


        -- =========================
-- Atome
-- 7,858.86
-- =========================

        INSERT INTO transactions(user_id,
                                 description)
        VALUES (v_user_id,
                'Opening Balance - MariBank')
        RETURNING id INTO tx;


        INSERT INTO entries(transaction_id,
                            account_id,
                            amount,
                            type)
        VALUES (tx,
                equity_account,
                7858.86,
                'DEBIT'),
               (tx,
                atome,
                7858.86,
                'CREDIT');


        -- =========================
-- SPayLater
-- 5,938.80
-- =========================

        INSERT INTO transactions(user_id,
                                 description)
        VALUES (v_user_id,
                'Opening Balance - SPayLater')
        RETURNING id INTO tx;


        INSERT INTO entries(transaction_id,
                            account_id,
                            amount,
                            type)
        VALUES (tx,
                equity_account,
                5938.80,
                'DEBIT'),
               (tx,
                spaylater,
                5938.80,
                'CREDIT');

    END
$$;
