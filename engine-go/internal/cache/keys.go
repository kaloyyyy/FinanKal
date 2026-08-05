package cache

import (
	"fmt"

	"github.com/google/uuid"
)

func UserCreditCardsKey(userID uuid.UUID) string {
	return fmt.Sprintf(
		"user:%s:credit_cards",
		userID,
	)
}

func CreditCardKey(cardID uuid.UUID) string {
	return fmt.Sprintf(
		"credit_card:%s",
		cardID,
	)
}

func CreditCardStatementKey(statementID uuid.UUID) string {
	return fmt.Sprintf(
		"credit_card_statement:%s",
		statementID,
	)
}

func AccountBalanceKey(accountID uuid.UUID) string {
	return fmt.Sprintf(
		"account:%s:balance",
		accountID,
	)
}

func AccountSummaryKey(accountID uuid.UUID) string {
	return fmt.Sprintf(
		"account:%s:summary",
		accountID,
	)
}

func AccountLedgerEntriesKey(accountID uuid.UUID) string {
	return fmt.Sprintf(
		"account:%s:ledger",
		accountID,
	)
}

func UserBalanceKey(userID uuid.UUID) string {
	return fmt.Sprintf(
		"user:%s:balance",
		userID,
	)
}

func UserTotalCreditKey(userID uuid.UUID) string {
	return fmt.Sprintf(
		"user:%s:total_credit",
		userID,
	)
}

func UserTotalDebitKey(userID uuid.UUID) string {
	return fmt.Sprintf(
		"user:%s:total_debit",
		userID,
	)
}

func UserTotalBalanceKey(userID uuid.UUID) string {
	return fmt.Sprintf(
		"user:%s:total_balance",
		userID,
	)
}
