package credit

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EntryRepository struct {
	db *pgxpool.Pool
}

func NewEntryRepository(db *pgxpool.Pool) *EntryRepository {
	return &EntryRepository{db: db}
}

func (r *EntryRepository) CreateEntryMappingTx(
	ctx context.Context,
	tx pgx.Tx,
	statementID uuid.UUID,
	entryID uuid.UUID,
) error {

	_, err := tx.Exec(
		ctx,
		`
		INSERT INTO credit_card_entries (
			statement_id,
			entry_id
		)
		VALUES ($1,$2)
		`,
		statementID,
		entryID,
	)

	return err
}
