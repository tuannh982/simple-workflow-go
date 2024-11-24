//go:build e2e
// +build e2e

package bank_transfer

import (
	"context"
	"fmt"
	"github.com/tuannh982/simple-workflow-go/pkg/api/workflow"
)

type Void struct{}

type PaymentWorkflowInput struct {
	FromBank    string `json:"from_bank"`
	FromAccount string `json:"from_account"`
	ToBank      string `json:"to_bank"`
	ToAccount   string `json:"to_account"`
	Amount      int    `json:"amount"`
}

type InterBankTransferActivityInput struct {
	Bank   string `json:"bank"`
	From   string `json:"from"`
	To     string `json:"to"`
	Amount int    `json:"amount"`
}

func (i *InterBankTransferActivityInput) Reverse() *InterBankTransferActivityInput {
	return &InterBankTransferActivityInput{
		Bank:   i.Bank,
		From:   i.To,
		To:     i.From,
		Amount: i.Amount,
	}
}

type CrossBankTransferInput struct {
	FromBank string `json:"from_bank"`
	ToBank   string `json:"to_bank"`
	Amount   int    `json:"amount"`
}

func (c *CrossBankTransferInput) Reverse() *CrossBankTransferInput {
	return &CrossBankTransferInput{
		FromBank: c.ToBank,
		ToBank:   c.FromBank,
		Amount:   c.Amount,
	}
}

var mockPaymentDB *MockPaymentDB

func InterBankTransferActivity(ctx context.Context, input *InterBankTransferActivityInput) (*Void, error) {
	if bank, ok := mockPaymentDB.Banks[input.Bank]; ok {
		if err := bank.Transfer(input.From, input.To, input.Amount); err != nil {
			return nil, err
		}
		return &Void{}, nil
	} else {
		return nil, fmt.Errorf("bank %s does not exist", input.Bank)
	}
}

func CrossBankTransferActivity(ctx context.Context, input *CrossBankTransferInput) (*Void, error) {
	if err := mockPaymentDB.Transfer(input.FromBank, input.ToBank, input.Amount); err != nil {
		return nil, err
	}
	return &Void{}, nil
}

func PaymentWorkflow(ctx context.Context, input *PaymentWorkflowInput) (result *Void, err error) {
	if input.Amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than zero")
	}
	if input.FromBank == input.ToBank {
		txInput := &InterBankTransferActivityInput{
			Bank:   input.FromBank,
			From:   input.FromAccount,
			To:     input.ToAccount,
			Amount: input.Amount,
		}
		_, err = workflow.CallActivity(ctx, InterBankTransferActivity, txInput).Await()
		if err != nil {
			return nil, err
		}
		defer func() {
			if err != nil {
				_, err = workflow.CallActivity(ctx, InterBankTransferActivity, txInput.Reverse()).Await()
			}
		}()
	} else {
		// transfer from account A to bank's reserved account
		tx1Input := &InterBankTransferActivityInput{
			Bank:   input.FromBank,
			From:   input.FromAccount,
			To:     ReservedAccountID,
			Amount: input.Amount,
		}
		// transfer between 2 banks reserved accounts
		tx2Input := &CrossBankTransferInput{
			FromBank: input.FromBank,
			ToBank:   input.ToBank,
			Amount:   input.Amount,
		}
		// transfer from bank's reserved account to account B
		tx3Input := &InterBankTransferActivityInput{
			Bank:   input.ToBank,
			From:   ReservedAccountID,
			To:     input.ToAccount,
			Amount: input.Amount,
		}
		_, err = workflow.CallActivity(ctx, InterBankTransferActivity, tx1Input).Await()
		if err != nil {
			return nil, err
		}
		defer func() {
			if err != nil {
				_, _ = workflow.CallActivity(ctx, InterBankTransferActivity, tx1Input.Reverse()).Await()
			}
		}()
		_, err = workflow.CallActivity(ctx, CrossBankTransferActivity, tx2Input).Await()
		if err != nil {
			return nil, err
		}
		defer func() {
			if err != nil {
				_, _ = workflow.CallActivity(ctx, CrossBankTransferActivity, tx2Input.Reverse()).Await()
			}
		}()
		_, err = workflow.CallActivity(ctx, InterBankTransferActivity, tx3Input).Await()
		if err != nil {
			return nil, err
		}
		defer func() {
			if err != nil {
				_, _ = workflow.CallActivity(ctx, InterBankTransferActivity, tx3Input.Reverse()).Await()
			}
		}()
	}
	return &Void{}, nil
}
