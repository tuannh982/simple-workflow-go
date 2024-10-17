package history

type HistoryEvent struct {
	EventID   int32
	Timestamp int64
	Event
}
