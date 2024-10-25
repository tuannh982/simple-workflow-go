package history

type TimerCreated struct {
	TimerID int64
	FireAt  int64
}

type TimerFired struct {
	TimerID int64
}
