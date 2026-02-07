package enqueuer

import (
	"log"
)

type EmptyRequestEnqueuer struct {
	logger *log.Logger
}

func NewEmptyRequestEnqueuer(logger *log.Logger) *EmptyRequestEnqueuer {
	return &EmptyRequestEnqueuer{logger: logger}
}

func (e *EmptyRequestEnqueuer) EnqueueRequest(request Request) error {
	if e == nil || e.logger == nil {
		return nil
	}
	e.logger.Printf("[enqueue] name=%s fields=%v", request.Name, request.Fields)
	return nil
}
