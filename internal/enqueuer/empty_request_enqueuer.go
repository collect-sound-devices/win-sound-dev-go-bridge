package enqueuer

import "github.com/collect-sound-devices/win-sound-dev-go-bridge/internal/logging"

type EmptyRequestEnqueuer struct {
	logger logging.Logger
}

func NewEmptyRequestEnqueuer(logger logging.Logger) *EmptyRequestEnqueuer {
	return &EmptyRequestEnqueuer{logger: logger}
}

func (e *EmptyRequestEnqueuer) EnqueueRequest(request Request) error {
	if e == nil || e.logger == nil {
		return nil
	}
	e.logger.Printf("[info, empty enqueuer] name=%s fields=%v", request.Name, request.Fields)
	return nil
}
