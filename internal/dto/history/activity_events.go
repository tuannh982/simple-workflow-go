package history

import "github.com/tuannh982/simple-workflows-go/internal/dto"

type ActivityScheduled struct {
	TaskScheduledID int32
	Name            string
	Input           []byte
}

type ActivityCompleted struct {
	TaskScheduledID int32
	dto.ExecutionResult
}
