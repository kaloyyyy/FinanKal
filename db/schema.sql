create table users
(
    id         uuid                     default gen_random_uuid() not null
        primary key,
    name       varchar(255)                                       not null,
    username   varchar(255)                                       not null
        unique,
    created_at timestamp with time zone default now()
);

alter table users
    owner to postgres;

create table accounts
(
    id         uuid                     default gen_random_uuid() not null
        primary key,
    user_id    uuid                                               not null
        references users
            on delete cascade,
    name       varchar(255)                                       not null,
    type       varchar(50)                                        not null
        constraint chk_account_type
            check ((type)::text = ANY
                   ((ARRAY ['ASSET'::character varying, 'LIABILITY'::character varying, 'EXPENSE'::character varying, 'INCOME'::character varying, 'EQUITY'::character varying, 'CREDIT_CARD'::character varying, 'CLEARING'::character varying])::text[])),
    created_at timestamp with time zone default now()
);

alter table accounts
    owner to postgres;

create index idx_accounts_user_id
    on accounts (user_id);

create table transactions
(
    id          uuid                     default gen_random_uuid() not null
        primary key,
    description text,
    created_at  timestamp with time zone default now(),
    user_id     uuid
        references users
            on delete cascade
);

alter table transactions
    owner to postgres;

create index idx_transactions_user_id
    on transactions (user_id);

create table entries
(
    id             uuid                     default gen_random_uuid() not null
        primary key,
    transaction_id uuid                                               not null
        references transactions
            on delete cascade,
    account_id     uuid                                               not null
        references accounts
            on delete cascade,
    amount         numeric(19, 4)                                     not null,
    type           varchar(10)                                        not null
        constraint entries_type_check
            check ((type)::text = ANY ((ARRAY ['DEBIT'::character varying, 'CREDIT'::character varying])::text[])),
    created_at     timestamp with time zone default now(),
    statement_id   uuid
);

alter table entries
    owner to postgres;

create index idx_entries_account_id
    on entries (account_id);

create index idx_entries_transaction_id
    on entries (transaction_id);

create index idx_entries_account
    on entries (account_id);

create table credit_cards
(
    id                  uuid                                   not null
        primary key
        references accounts
            on delete cascade,
    clearing_account_id uuid                                   not null
        references accounts,
    credit_limit        numeric(18, 2)                         not null
        constraint credit_cards_credit_limit_check
            check (credit_limit >= (0)::numeric),
    billing_day         integer                                not null
        constraint credit_cards_billing_day_check
            check ((billing_day >= 1) AND (billing_day <= 28)),
    payment_due_days    integer                                not null
        constraint credit_cards_payment_due_days_check
            check ((payment_due_days >= 1) AND (payment_due_days <= 28)),
    created_at          timestamp with time zone default now() not null,
    updated_at          timestamp with time zone default now() not null
);

alter table credit_cards
    owner to postgres;

create index idx_credit_cards_clearing_account
    on credit_cards (clearing_account_id);

create table credit_card_statements
(
    id             uuid                     default gen_random_uuid() not null
        primary key,
    credit_card_id uuid                                               not null
        references credit_cards
            on delete cascade,
    start_date     date                                               not null,
    end_date       date                                               not null,
    statement_date date                                               not null,
    due_date       date                                               not null,
    total_amount   numeric(18, 2)           default 0                 not null,
    status         varchar(20)                                        not null,
    created_at     timestamp with time zone default now()             not null,
    updated_at     timestamp with time zone default now()             not null
);

alter table credit_card_statements
    owner to postgres;

create index idx_credit_card_statements_card
    on credit_card_statements (credit_card_id);

create table credit_card_entries
(
    statement_id uuid not null
        references credit_card_statements
            on delete cascade,
    entry_id     uuid not null
        references entries
            on delete cascade,
    primary key (statement_id, entry_id)
);

alter table credit_card_entries
    owner to postgres;

create index idx_credit_card_entries_statement
    on credit_card_entries (statement_id);

create index idx_credit_card_entries_entry
    on credit_card_entries (entry_id);

-- ADR-0001: bank statements as stored records (same as db/migrations/001_bank_statements.sql).
create table bank_statements
(
    id              uuid                     default gen_random_uuid() not null
        primary key,
    account_id      uuid                                               not null
        references accounts
            on delete cascade,
    bank            varchar(20)                                        not null,
    period_start    date                                               not null,
    period_end      date                                               not null,
    opening_balance numeric(19, 4),
    closing_balance numeric(19, 4),
    file_sha256     char(64)                                           not null,
    imported_at     timestamp with time zone default now()             not null,
    constraint bank_statements_period_check
        check (period_end >= period_start),
    constraint bank_statements_file_unique
        unique (account_id, file_sha256),
    constraint bank_statements_period_unique
        unique (account_id, period_start, period_end)
);

alter table bank_statements
    owner to postgres;

create index idx_bank_statements_account
    on bank_statements (account_id, period_start);

alter table transactions
    add column statement_id uuid
        references bank_statements
            on delete cascade,
    add column opening_balance boolean default false not null;

create index idx_transactions_statement_id
    on transactions (statement_id);


-- 002_account_numbers.sql
alter table accounts
    add column account_number varchar(34);

create unique index accounts_user_account_number_unique
    on accounts (user_id, account_number)
    where account_number is not null;

-- 003_card_bills.sql
alter table credit_cards
    add column due_rule varchar(20) default 'DAYS' not null,
    add constraint credit_cards_due_rule_check check (due_rule in ('DAYS', 'DAY_OF_MONTH', 'BUSINESS_DAYS'));

alter table credit_card_statements
    alter column due_date drop not null;

create unique index credit_card_statements_card_date_unique
    on credit_card_statements (credit_card_id, statement_date);

-- 004_statement_shift.sql
alter table credit_cards
    add column statement_shift varchar(25) default 'NEXT_BUSINESS_DAY' not null,
    add constraint credit_cards_statement_shift_check
        check (statement_shift in ('NONE', 'NEXT_BUSINESS_DAY', 'NEAREST_BUSINESS_DAY'));

-- 005_loans.sql
create table loans
(
    account_id   uuid primary key references accounts (id) on delete cascade,
    booked_on    date           not null,
    daily_rate   numeric(10, 8) not null check (daily_rate >= 0),
    principal_on date           not null,
    principal    numeric(19, 4) not null,
    file_sha256  char(64)       not null unique,
    paid_through date
);

create table loan_charges
(
    transaction_id uuid primary key references transactions (id) on delete cascade,
    account_id     uuid not null references accounts (id) on delete cascade
);

create index loan_charges_account_id
    on loan_charges (account_id);
