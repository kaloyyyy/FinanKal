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

func AccountBalanceKey(accountID uuid.UUID) string {
	return fmt.Sprintf("account:%s:balance", accountID)
}

func AccountSummaryKey(accountID uuid.UUID) string {
	return fmt.Sprintf("account:%s:summary", accountID)
}

func AccountLedgerEntriesKey(accountID uuid.UUID) string {
	return fmt.Sprintf("account:%s:ledger", accountID)
}

func UserTotalCreditKey(userID uuid.UUID) string {
	return fmt.Sprintf("user:%s:total_credit", userID)
}

func UserTotalDebitKey(userID uuid.UUID) string {
	return fmt.Sprintf("user:%s:total_debit", userID)
}

func UserNetWorthKey(userID uuid.UUID) string {
	return fmt.Sprintf("user:%s:total_balance", userID)
}

func UserAccountsKey(userID uuid.UUID) string {
	return fmt.Sprintf("user:%s:accounts", userID)
}
