package dto

type Error struct {
	Type        string
	Message     string
	IsRetryable bool
}
