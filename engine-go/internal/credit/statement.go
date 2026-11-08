package credit

import (
	"context"

	"github.com/google/uuid"
)

// GetStatement is deliberately uncached: it is read before paying, so it must not be stale.
func (s *Service) GetStatement(ctx context.Context, statementID uuid.UUID) (*CreditCardStatement, error) {
	return s.statementRepo.GetStatement(ctx, statementID)
}
