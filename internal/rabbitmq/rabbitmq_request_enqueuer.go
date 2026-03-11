package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/collect-sound-devices/win-sound-dev-go-bridge/internal/contract"
	"github.com/collect-sound-devices/win-sound-dev-go-bridge/internal/enqueuer"
	"github.com/collect-sound-devices/win-sound-dev-go-bridge/internal/logging"
)

// RabbitMessagePublisher is the publish contract expected from a RabbitMQ publisher.
type RabbitMessagePublisher interface {
	Publish(ctx context.Context, body []byte) error
	Close() error
}

// RabbitMqEnqueuer writes requests to RabbitMQ using the same message-shaping
type RabbitMqEnqueuer struct {
	baseCtx        context.Context
	publisher      RabbitMessagePublisher
	logger         logging.Logger
	publishTimeout time.Duration
}

func NewRabbitMqEnqueuerWithContext(baseCtx context.Context, publisher RabbitMessagePublisher, logger logging.Logger) *RabbitMqEnqueuer {
	if baseCtx == nil {
		panic("nil context")
	}
	if publisher == nil {
		panic("nil publisher")
	}
	if logger == nil {
		panic("nil logger")
	}

	return newRabbitMqEnqueuer(
		baseCtx,
		publisher,
		logger,
		10*time.Second,
	)
}

func newRabbitMqEnqueuer(
	baseCtx context.Context,
	publisher RabbitMessagePublisher,
	logger logging.Logger,
	publishTimeout time.Duration,
) *RabbitMqEnqueuer {
	if publishTimeout <= 0 {
		publishTimeout = 10 * time.Second
	}

	return &RabbitMqEnqueuer{
		baseCtx:        baseCtx,
		publisher:      publisher,
		logger:         logger,
		publishTimeout: publishTimeout,
	}
}

func (e *RabbitMqEnqueuer) EnqueueRequest(request enqueuer.Request) error {
	payload := make(map[string]any, len(request.Fields)+4)
	for key, value := range request.Fields {
		payload[key] = normalizeValue(key, value)
	}

	flowType, messageType := calculateFlowAndMessageType(request.Event)
	payload[contract.FieldDeviceMessageType] = messageType

	httpRequest, urlSuffix := e.resolveHttpRequest(request, payload)
	payload[contract.FieldHTTPRequest] = httpRequest
	payload[contract.FieldURLSuffix] = urlSuffix

	if httpRequest == "POST" {
		if flowType != 0 {
			payload[contract.FieldFlowType] = flowType
		}
	}
	if _, ok := payload[contract.FieldUpdateDate]; !ok && !request.Timestamp.IsZero() {
		payload[contract.FieldUpdateDate] = request.Timestamp.UTC().Format(time.RFC3339)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal rabbitmq payload: %w", err)
	}

	e.logf("[info, rabbitmq enqueuer] publishing method=%s urlSuffix=%s", httpRequest, urlSuffix)

	ctx, cancel := context.WithTimeout(e.baseCtx, e.publishTimeout)
	defer cancel()
	if err := e.publisher.Publish(ctx, body); err != nil {
		return fmt.Errorf("publish request: %w", err)
	}

	return nil
}

func (e *RabbitMqEnqueuer) Close() error {
	return e.publisher.Close()
}

func (e *RabbitMqEnqueuer) resolveHttpRequest(request enqueuer.Request, payload map[string]any) (string, string) {
	var httpRequest string

	switch request.Event {
	case contract.EventTypeRenderDeviceDiscovered,
		contract.EventTypeCaptureDeviceDiscovered,
		contract.EventTypeRenderDeviceConfirmed,
		contract.EventTypeCaptureDeviceConfirmed:
		httpRequest = "POST"
	default:
		httpRequest = "PUT"
	}

	urlSuffix := readStringField(payload, contract.FieldURLSuffix)
	if urlSuffix == "" && httpRequest == "PUT" {
		pnpID := readStringField(payload, contract.FieldPnpID)
		hostName := readStringField(payload, contract.FieldHostName)

		urlSuffix = fmt.Sprintf("/%s/%s", pnpID, hostName)
		delete(payload, contract.FieldHostName) // contract.FieldHostName is only used for building the URL suffix, so we must remove it from the payload
	}

	return httpRequest, urlSuffix
}

func readStringField(payload map[string]any, key string) string {
	if v, ok := payload[key]; ok {
		s, okString := v.(string)
		if okString {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

func calculateFlowAndMessageType(event contract.EventType) (contract.FlowType, contract.MessageType) {

	var flow contract.FlowType
	var message contract.MessageType

	switch event {
	case contract.EventTypeRenderDeviceConfirmed, contract.EventTypeRenderDeviceDiscovered, contract.EventTypeRenderVolumeChanged:
		flow = contract.FlowTypeRender
	case contract.EventTypeCaptureDeviceConfirmed, contract.EventTypeCaptureDeviceDiscovered, contract.EventTypeCaptureVolumeChanged:
		flow = contract.FlowTypeCapture
	default:
		flow = 0
	}

	switch event {
	case contract.EventTypeRenderDeviceConfirmed, contract.EventTypeCaptureDeviceConfirmed:
		message = contract.MessageTypeConfirmed
	case contract.EventTypeRenderDeviceDiscovered, contract.EventTypeCaptureDeviceDiscovered:
		message = contract.MessageTypeDiscovered
	case contract.EventTypeRenderVolumeChanged:
		message = contract.MessageTypeVolumeRenderChanged
	case contract.EventTypeCaptureVolumeChanged:
		message = contract.MessageTypeVolumeCaptureChanged
	default:
		message = 0
	}

	return flow, message
}

func normalizeValue(key string, value string) any {
	trimmed := strings.TrimSpace(value)
	switch key {
	case contract.FieldRenderVolume, contract.FieldCaptureVolume, contract.FieldVolume:
		if n, err := strconv.Atoi(trimmed); err == nil {
			return n
		}
	}
	return value
}

func (e *RabbitMqEnqueuer) logf(format string, args ...interface{}) {
	e.logger.Printf(format, args...)
}
