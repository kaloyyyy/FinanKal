package main

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/kaloy/finankal/engine-go/finance"
	"github.com/kaloy/finankal/engine-go/internal/credit"
	"github.com/kaloy/finankal/engine-go/internal/ledger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestToStatus(t *testing.T) {
	tests := []struct {
		err  error
		code codes.Code
	}{
		{pgx.ErrNoRows, codes.NotFound},
		{fmt.Errorf("%w: unbalanced", ledger.ErrInvalidTransaction), codes.InvalidArgument},
		{credit.ErrInvalidAmount, codes.InvalidArgument},
		{credit.ErrStatementCardMismatch, codes.InvalidArgument},
		{credit.ErrCreditLimitExceeded, codes.FailedPrecondition},
		{credit.ErrStatementAlreadyPaid, codes.FailedPrecondition},
		{credit.ErrPaymentTooLarge, codes.FailedPrecondition},
	}

	for _, test := range tests {
		if got := status.Code(toStatus(test.err)); got != test.code {
			t.Errorf("%v: got %s, want %s", test.err, got, test.code)
		}
	}
}

func TestToStatusHidesUnexpectedErrors(t *testing.T) {
	st, _ := status.FromError(toStatus(errors.New(`relation "accounts" does not exist`)))

	if st.Code() != codes.Internal || st.Message() != "internal error" {
		t.Errorf("got %s %q", st.Code(), st.Message())
	}
}

func TestClampLimit(t *testing.T) {
	tests := []struct {
		limit int32
		want  int
	}{{0, 5}, {-1, 5}, {3, 3}, {20, 20}, {21, 20}}

	for _, test := range tests {
		if got := clampLimit(test.limit, 5, 20); got != test.want {
			t.Errorf("clampLimit(%d): got %d, want %d", test.limit, got, test.want)
		}
	}
}

func TestRecordCreditCardTransactionRequiresPurchaseDate(t *testing.T) {
	_, err := (&Server{}).RecordCreditCardTransaction(context.Background(), &finance.RecordCreditCardTransactionRequest{
		CardId: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		Amount: "10",
	})

	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("got %v, want InvalidArgument", err)
	}
}
