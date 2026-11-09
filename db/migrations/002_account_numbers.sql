-- The account number printed on a bank statement, so imports can find the account by themselves.
BEGIN;

ALTER TABLE accounts ADD COLUMN IF NOT EXISTS account_number varchar(34);

CREATE UNIQUE INDEX IF NOT EXISTS accounts_user_account_number_unique
    ON accounts (user_id, account_number)
    WHERE account_number IS NOT NULL;

COMMIT;
