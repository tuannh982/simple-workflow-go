package subscription_with_debug

import "time"

func calculatePaymentCycles(totalAmount int64, cycles int) []int64 {
	div := totalAmount / int64(cycles)
	mod := int(totalAmount % int64(cycles))
	paymentAmounts := make([]int64, cycles)
	for i := 0; i < cycles; i++ {
		paymentAmounts[i] = div
		if i < mod {
			paymentAmounts[i] += 1
		}
	}
	return paymentAmounts
}

func calculatePaymentTimings(anchor int64, cycles int, cycleDuration time.Duration) []int64 {
	timings := make([]int64, cycles)
	for i := 0; i < cycles; i++ {
		d := time.Duration(i+1) * cycleDuration
		timings[i] = anchor + d.Milliseconds()
	}
	return timings
}
