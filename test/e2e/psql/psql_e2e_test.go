package psql

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tuannh982/simple-workflows-go/pkg/api/client"
	"go.uber.org/zap"
	"testing"
)

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func InitMocks() {
	mockPaymentDB = NewMockPaymentDB()
	bankA := NewMockBank("A")
	bankB := NewMockBank("B")
	Must(mockPaymentDB.AddBank(bankA))
	Must(mockPaymentDB.AddBank(bankB))
	Must(bankA.AddAccount("A_1", 1000))
	Must(bankA.AddAccount("A_2", 1000))
	Must(bankB.AddAccount("B_1", 1000))
	Must(bankB.AddAccount("B_2", 1000))
	bankB.InjectTransferError(func(from string, to string, amount int) error {
		if to == "B_2" {
			return errors.New("account locked")
		}
		return nil
	})
}

func InitLogger() (*zap.Logger, error) {
	return zap.NewProduction()
}

func Test(t *testing.T) {
	InitMocks()
	ctx := context.Background()
	logger, err := InitLogger()
	assert.NoError(t, err)
	be, err := InitBackend()
	assert.NoError(t, err)
	aw, ww, err := InitWorkers(be, logger)
	assert.NoError(t, err)
	aw.Start(ctx)
	defer aw.Stop(ctx)
	ww.Start(ctx)
	defer ww.Stop(ctx)
	//
	uid, err := uuid.NewV7()
	assert.NoError(t, err)
	workflowID := fmt.Sprintf("e2e-test-payment-workflow-%s", uid.String())
	err = client.ScheduleWorkflow(ctx, be, PaymentWorkflow, &PaymentWorkflowInput{
		FromBank:    "A",
		FromAccount: "A_1",
		ToBank:      "B",
		ToAccount:   "B_2",
		Amount:      800,
	}, client.WorkflowScheduleOptions{
		WorkflowID: workflowID,
		Version:    "1",
	})
	assert.NoError(t, err)
	wResult, wErr, err := client.AwaitWorkflowResult(ctx, be, PaymentWorkflow, workflowID)
	assert.Nil(t, wResult)
	assert.Error(t, wErr)
	assert.Equal(t, "account locked", wErr.Error())
}
