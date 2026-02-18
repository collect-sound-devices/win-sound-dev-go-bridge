package enqueuer

import (
	"github.com/collect-sound-devices/win-sound-dev-go-bridge/internal/logging"
)

type EmptyRequestEnqueuer struct {
	logger logging.Logger
}

func NewEmptyRequestEnqueuer(logger logging.Logger) *EmptyRequestEnqueuer {
	if logger == nil {
		panic("nil logger")
	}
	return &EmptyRequestEnqueuer{logger: logger}
}

func (e *EmptyRequestEnqueuer) EnqueueRequest(request Request) error {
	e.logger.Printf(
		"[info, empty enqueuer] event=%d fields=%v",
		request.Event,
		request.Fields,
	)
	return nil
}
