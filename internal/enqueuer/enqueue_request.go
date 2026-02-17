package enqueuer

import "time"

type Request struct {
	Timestamp   time.Time
	MessageType uint8
	Fields      map[string]string
}

type EnqueueRequest interface {
	EnqueueRequest(request Request) error
}
