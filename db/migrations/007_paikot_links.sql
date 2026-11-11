-- Paikot: a due on account_id can be paid with card_id (PayDirect, QRph), which moves it onto that
-- card's next bill. Forecast only: nothing is booked from these links.
BEGIN;

CREATE TABLE IF NOT EXISTS paikot_links (
    account_id  uuid PRIMARY KEY REFERENCES accounts (id) ON DELETE CASCADE,
    card_id     uuid          NOT NULL REFERENCES credit_cards (id) ON DELETE CASCADE,
    fee_percent numeric(6, 3) NOT NULL DEFAULT 0 CHECK (fee_percent BETWEEN 0 AND 100),
    CHECK (account_id <> card_id)
);

COMMIT;
