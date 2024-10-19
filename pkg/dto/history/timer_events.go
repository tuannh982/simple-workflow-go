package history

type TimerCreated struct {
	TimerID int32
	FireAt  int64
}

type TimerFired struct {
	TimerID int32
}
