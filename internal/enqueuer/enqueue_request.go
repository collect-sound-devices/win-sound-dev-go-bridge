package enqueuer

import (
	"github.com/collect-sound-devices/win-sound-dev-go-bridge/internal/contract"
	"time"
)

type Request struct {
	Timestamp time.Time
	Event     contract.EventType
	Fields    map[string]string
}

type EnqueueRequest interface {
	EnqueueRequest(request Request) error
}
