package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
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
	baseCtx          context.Context
	publisher        RabbitMessagePublisher
	logger           logging.Logger
	hostName         string
	operationSysName string
	publishTimeout   time.Duration
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

	hostName, err := os.Hostname()
	if err != nil || strings.TrimSpace(hostName) == "" {
		hostName = "unknown-host"
	}
	return newRabbitMqEnqueuer(
		baseCtx,
		publisher,
		logger,
		hostName,
		runtime.GOOS,
		10*time.Second,
	)
}

func newRabbitMqEnqueuer(
	baseCtx context.Context,
	publisher RabbitMessagePublisher,
	logger logging.Logger,
	hostName string,
	operationSysName string,
	publishTimeout time.Duration,
) *RabbitMqEnqueuer {
	if strings.TrimSpace(hostName) == "" {
		hostName = "unknown-host"
	}
	if strings.TrimSpace(operationSysName) == "" {
		operationSysName = runtime.GOOS
	}
	if publishTimeout <= 0 {
		publishTimeout = 10 * time.Second
	}

	return &RabbitMqEnqueuer{
		baseCtx:          baseCtx,
		publisher:        publisher,
		logger:           logger,
		hostName:         hostName,
		operationSysName: operationSysName,
		publishTimeout:   publishTimeout,
	}
}

func (e *RabbitMqEnqueuer) EnqueueRequest(request enqueuer.Request) error {
	payload := make(map[string]any, len(request.Fields)+4)
	for key, value := range request.Fields {
		payload[key] = normalizeValue(key, value)
	}
	payload[contract.FieldDeviceMessageType] = request.MessageType

	httpRequest, urlSuffix := e.resolveHttpRequest(request, payload)
	payload[contract.FieldHTTPRequest] = httpRequest
	payload[contract.FieldURLSuffix] = urlSuffix

	if httpRequest == "POST" {
		if _, ok := payload[contract.FieldHostName]; !ok {
			payload[contract.FieldHostName] = e.hostName
		}
		if _, ok := payload[contract.FieldOperationSystemName]; !ok {
			payload[contract.FieldOperationSystemName] = e.operationSysName
		}

		flowType := calculateFlowTypeField(contract.MessageType(request.MessageType))
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
	messageType := contract.MessageType(request.MessageType)
	var httpRequest string
	switch messageType {
	case contract.MessageTypeDefaultRenderChanged, contract.MessageTypeDefaultCaptureChanged:
		httpRequest = "POST"
	default:
		httpRequest = "PUT"
	}

	urlSuffix := readStringField(payload, contract.FieldURLSuffix)
	if urlSuffix == "" && httpRequest == "PUT" {
		pnpID := readStringField(payload, contract.FieldPnpID)
		urlSuffix = fmt.Sprintf("/%s/%s", pnpID, e.hostName)
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

func calculateFlowTypeField(messageType contract.MessageType) contract.FlowType {
	switch messageType {
	case contract.MessageTypeDefaultRenderChanged, contract.MessageTypeVolumeRenderChanged:
		return contract.FlowTypeRender
	case contract.MessageTypeDefaultCaptureChanged, contract.MessageTypeVolumeCaptureChanged:
		return contract.FlowTypeCapture
	}
	return 0
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
