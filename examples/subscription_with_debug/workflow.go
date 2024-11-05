package subscription_with_debug

import (
	"context"
	"errors"
	"fmt"
	"github.com/tuannh982/simple-workflows-go/pkg/api/workflow"
	"math/rand"
	"time"
)

type Void struct {
}

type SubscriptionWorkflowInput struct {
	TotalAmount   int64
	Cycles        int
	CycleDuration time.Duration
}

type SubscriptionWorkflowOutput struct {
	Paid    int64
	Overdue int64
}

type PaymentInput struct {
	Value int64
}

func PaymentActivity(ctx context.Context, input *PaymentInput) (*Void, error) {
	r := rand.Intn(100)
	if r < 30 { // 30% of failure
		return &Void{}, nil
	} else {
		return nil, errors.New("payment failed")
	}
}

func SubscriptionWorkflow(ctx context.Context, input *SubscriptionWorkflowInput) (*SubscriptionWorkflowOutput, error) {
	startTimestamp := workflow.GetWorkflowExecutionStartedTimestamp(ctx)
	paymentAmounts := calculatePaymentCycles(input.TotalAmount, input.Cycles)
	paymentTimings := calculatePaymentTimings(startTimestamp, input.Cycles, input.CycleDuration)
	//
	var paid int64 = 0
	var overdue int64 = 0
	currentCycle := 0
	for {
		workflow.SetVar(ctx, "paid", paid)
		workflow.SetVar(ctx, "overdue", overdue)
		workflow.SetVar(ctx, "currentCycle", currentCycle)
		if currentCycle >= input.Cycles {
			break
		}
		currentCycleAmount := paymentAmounts[currentCycle]
		amountToPay := currentCycleAmount + overdue
		workflow.SetVar(ctx, "amountToPay", amountToPay)
		workflow.WaitUntil(ctx, time.UnixMilli(paymentTimings[currentCycle]))
		_, err := workflow.CallActivity(ctx, PaymentActivity, &PaymentInput{Value: amountToPay}).Await()
		if err != nil {
			overdue += paymentAmounts[currentCycle]
			workflow.SetVar(ctx, fmt.Sprintf("cycle_%d_err", currentCycle), err.Error())
		} else {
			paid += amountToPay
			overdue = 0
			workflow.SetVar(ctx, fmt.Sprintf("cycle_%d_paid_amount", currentCycle), amountToPay)
		}
		workflow.SetVar(ctx, "amountToPay", 0)
		workflow.SetVar(ctx, fmt.Sprintf("cycle_%d_completed_at", currentCycle), workflow.GetCurrentTimestamp(ctx))
		currentCycle += 1
	}
	return &SubscriptionWorkflowOutput{
		Paid:    paid,
		Overdue: overdue,
	}, nil
}
