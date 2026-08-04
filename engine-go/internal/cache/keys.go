package cache

import (
	"fmt"

	"github.com/google/uuid"
)

func UserCreditCardsKey(userID uuid.UUID) string {
	return fmt.Sprintf("user:%s:credit_cards", userID)
}

func CreditCardKey(cardID uuid.UUID) string {
	return fmt.Sprintf("credit_card:%s", cardID)
}

func CreditCardStatementKey(statementID uuid.UUID) string {
	return fmt.Sprintf("credit_card_statement:%s", statementID)
}

func UserBalanceKey(userID uuid.UUID) string {
	return fmt.Sprintf("user:%s:balance", userID)
}

func AccountBalanceKey(accountID uuid.UUID) string {
	return fmt.Sprintf("account:%s:balance", accountID)
}
