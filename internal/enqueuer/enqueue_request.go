package enqueuer

import "time"

type Request struct {
	Name      string
	Timestamp time.Time
	Fields    map[string]string
}

type EnqueueRequest interface {
	EnqueueRequest(request Request) error
}
