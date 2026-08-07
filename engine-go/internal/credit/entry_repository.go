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
	return &EntryRepository{
		db: db,
	}
}

func (r *EntryRepository) CreateEntryMapping(
	ctx context.Context,
	statementID uuid.UUID,
	entryID uuid.UUID,
) (uuid.UUID, error) {

	var id uuid.UUID

	const query = `
		INSERT INTO credit_card_entries (
			statement_id,
			entry_id
		)
		VALUES (
			$1,
			$2
		)
		RETURNING id
		`

	err := r.db.QueryRow(
		ctx,
		query,
		statementID,
		entryID,
	).Scan(&id)

	return id, err
}

func (r *EntryRepository) GetEntryMapping(
	ctx context.Context,
	id uuid.UUID,
) (*CreditCardEntry, error) {

	var entry CreditCardEntry

	const query = `
		SELECT
			id,
			statement_id,
			entry_id,
			created_at
		FROM credit_card_entries
		WHERE id = $1
		`

	err := r.db.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&entry.ID,
		&entry.StatementID,
		&entry.EntryID,
		&entry.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &entry, nil
}

func (r *EntryRepository) ListEntriesByStatement(
	ctx context.Context,
	statementID uuid.UUID,
) ([]CreditCardEntry, error) {

	const query = `
		SELECT
			id,
			statement_id,
			entry_id,
			created_at
		FROM credit_card_entries
		WHERE statement_id = $1
		ORDER BY created_at
		`

	rows, err := r.db.Query(
		ctx,
		query,
		statementID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []CreditCardEntry

	for rows.Next() {
		var entry CreditCardEntry

		err := rows.Scan(
			&entry.ID,
			&entry.StatementID,
			&entry.EntryID,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

func (r *EntryRepository) DeleteEntryMapping(
	ctx context.Context,
	id uuid.UUID,
) error {

	const query = `
		DELETE
		FROM credit_card_entries
		WHERE id = $1
		`

	_, err := r.db.Exec(
		ctx,
		query,
		id,
	)

	return err
}

func (r *EntryRepository) DeleteMappingsByStatement(
	ctx context.Context,
	statementID uuid.UUID,
) error {

	const query = `
		DELETE
		FROM credit_card_entries
		WHERE statement_id = $1
		`

	_, err := r.db.Exec(
		ctx,
		query,
		statementID,
	)

	return err
}

func (r *EntryRepository) Exists(
	ctx context.Context,
	statementID uuid.UUID,
	entryID uuid.UUID,
) (bool, error) {

	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM credit_card_entries
			WHERE statement_id = $1
			  AND entry_id = $2
		)
		`

	var exists bool

	err := r.db.QueryRow(
		ctx,
		query,
		statementID,
		entryID,
	).Scan(&exists)

	return exists, err
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
