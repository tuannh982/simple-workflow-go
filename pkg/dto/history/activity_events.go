package history

import (
	"github.com/tuannh982/simple-workflows-go/pkg/dto"
)

type ActivityScheduled struct {
	TaskScheduledID int64
	Name            string
	Input           []byte
}

type ActivityCompleted struct {
	TaskScheduledID int64
	dto.ExecutionResult
}
