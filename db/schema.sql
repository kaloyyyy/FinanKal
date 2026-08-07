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

