package enqueuer

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

type testPublisher struct {
	lastBody []byte
	calls    int
	pubErr   error
}

func (p *testPublisher) Publish(_ context.Context, body []byte) error {
	p.calls++
	p.lastBody = append([]byte(nil), body...)
	return p.pubErr
}

func (p *testPublisher) Close() error {
	return nil
}

func TestRabbitMqEnqueuerPostDevice(t *testing.T) {
	pub := &testPublisher{}
	enq := newRabbitMqEnqueuer(pub, nil, "hostA", "windows", time.Second)

	err := enq.EnqueueRequest(Request{
		Name:      "post_device",
		Timestamp: time.Date(2026, 2, 10, 9, 8, 7, 0, time.UTC),
		Fields: map[string]string{
			"device_message_type": "default_render_changed",
			"update_date":         "2026-02-10T09:08:07Z",
			"flow_type":           "render",
			"name":                "Speaker",
			"pnp_id":              "PNP-1",
			"render_volume":       "32",
			"capture_volume":      "0",
		},
	})
	if err != nil {
		t.Fatalf("EnqueueRequest error: %v", err)
	}

	if pub.calls != 1 {
		t.Fatalf("expected 1 publish call, got %d", pub.calls)
	}

	var got map[string]any
	if err := json.Unmarshal(pub.lastBody, &got); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	if got["httpRequest"] != "POST" {
		t.Fatalf("expected httpRequest=POST, got %#v", got["httpRequest"])
	}
	if got["urlSuffix"] != "" {
		t.Fatalf("expected empty urlSuffix, got %#v", got["urlSuffix"])
	}
	if got["hostName"] != "hostA" {
		t.Fatalf("expected hostName=hostA, got %#v", got["hostName"])
	}
	if got["operationSystemName"] != "windows" {
		t.Fatalf("expected operationSystemName=windows, got %#v", got["operationSystemName"])
	}
	if got["renderVolume"] != float64(32) {
		t.Fatalf("expected renderVolume=32, got %#v", got["renderVolume"])
	}
	if got["captureVolume"] != float64(0) {
		t.Fatalf("expected captureVolume=0, got %#v", got["captureVolume"])
	}
}

func TestRabbitMqEnqueuerPutVolumeChange(t *testing.T) {
	pub := &testPublisher{}
	enq := newRabbitMqEnqueuer(pub, nil, "myHost", "windows", time.Second)

	err := enq.EnqueueRequest(Request{
		Name:      "put_volume_change",
		Timestamp: time.Date(2026, 2, 10, 9, 8, 7, 0, time.UTC),
		Fields: map[string]string{
			"device_message_type": "render_volume_changed",
			"volume":              "44",
			"pnp_id":              "DEV-22",
		},
	})
	if err != nil {
		t.Fatalf("EnqueueRequest error: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(pub.lastBody, &got); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	if got["httpRequest"] != "PUT" {
		t.Fatalf("expected httpRequest=PUT, got %#v", got["httpRequest"])
	}
	if got["urlSuffix"] != "/DEV-22/myHost" {
		t.Fatalf("expected urlSuffix=/DEV-22/myHost, got %#v", got["urlSuffix"])
	}
	if got["updateDate"] != "2026-02-10T09:08:07Z" {
		t.Fatalf("expected fallback updateDate from timestamp, got %#v", got["updateDate"])
	}
	if got["volume"] != float64(44) {
		t.Fatalf("expected volume=44, got %#v", got["volume"])
	}
}
