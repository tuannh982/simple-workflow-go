package bank_transfer

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tuannh982/simple-workflows-go/pkg/api/client"
	worker2 "github.com/tuannh982/simple-workflows-go/pkg/api/worker"
	"github.com/tuannh982/simple-workflows-go/pkg/utils/commons"
	"github.com/tuannh982/simple-workflows-go/test/e2e/psql"
	"go.uber.org/zap"
	"testing"
)

func InitMocks() {
	mockPaymentDB = NewMockPaymentDB()
	bankA := NewMockBank("A")
	bankB := NewMockBank("B")
	commons.Must(mockPaymentDB.AddBank(bankA))
	commons.Must(mockPaymentDB.AddBank(bankB))
	commons.Must(bankA.AddAccount("A_1", 1000))
	commons.Must(bankA.AddAccount("A_2", 1000))
	commons.Must(bankB.AddAccount("B_1", 1000))
	commons.Must(bankB.AddAccount("B_2", 1000))
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

// TODO FIXME sometime this test is flaky because the mock objects implementation is not deterministic
func Test(t *testing.T) {
	InitMocks()
	ctx := context.Background()
	logger, err := InitLogger()
	assert.NoError(t, err)
	be, err := psql.InitBackend(logger)
	assert.NoError(t, err)
	aw, err := worker2.NewActivityWorkersBuilder().
		WithName("[e2e test] ActivityWorker").
		WithBackend(be).
		WithLogger(logger).
		RegisterActivities(
			InterBankTransferActivity,
			CrossBankTransferActivity,
		).
		Build()
	assert.NoError(t, err)
	ww, err := worker2.NewWorkflowWorkersBuilder().
		WithName("[e2e test] WorkflowWorker").
		WithBackend(be).
		WithLogger(logger).
		RegisterWorkflows(
			PaymentWorkflow,
		).
		Build()
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
