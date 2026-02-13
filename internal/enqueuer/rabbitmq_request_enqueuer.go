package enqueuer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/collect-sound-devices/win-sound-dev-go-bridge/internal/logging"
)

const (
	fieldHttpRequest = "httpRequest"
	fieldURLSuffix   = "urlSuffix"
)

var snakeToCamelField = map[string]string{
	"device_message_type":  "deviceMessageType",
	"update_date":          "updateDate",
	"flow_type":            "flowType",
	"pnp_id":               "pnpId",
	"render_volume":        "renderVolume",
	"capture_volume":       "captureVolume",
	"host_name":            "hostName",
	"operation_system_name": "operationSystemName",
	"http_request":         fieldHttpRequest,
	"url_suffix":           fieldURLSuffix,
}

// RabbitMessagePublisher is the publish contract expected from a RabbitMQ publisher.
type RabbitMessagePublisher interface {
	Publish(ctx context.Context, body []byte) error
	Close() error
}

// RabbitMqEnqueuer writes requests to RabbitMQ using the same message-shaping
// approach as SoundWinAgent's RabbitMqHttpRequestDispatcher.
type RabbitMqEnqueuer struct {
	publisher        RabbitMessagePublisher
	logger           logging.Logger
	hostName         string
	operationSysName string
	publishTimeout   time.Duration
}

func NewRabbitMqEnqueuer(publisher RabbitMessagePublisher, logger logging.Logger) *RabbitMqEnqueuer {
	hostName, err := os.Hostname()
	if err != nil || strings.TrimSpace(hostName) == "" {
		hostName = "unknown-host"
	}
	return newRabbitMqEnqueuer(
		publisher,
		logger,
		hostName,
		runtime.GOOS,
		10*time.Second,
	)
}

func newRabbitMqEnqueuer(
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
		publisher:        publisher,
		logger:           logger,
		hostName:         hostName,
		operationSysName: operationSysName,
		publishTimeout:   publishTimeout,
	}
}

func (e *RabbitMqEnqueuer) EnqueueRequest(request Request) error {
	if e == nil || e.publisher == nil {
		return fmt.Errorf("rabbit mq enqueuer is not initialized")
	}

	payload := make(map[string]any, len(request.Fields)+4)
	for key, value := range request.Fields {
		mappedKey := mapFieldName(key)
		payload[mappedKey] = normalizeValue(mappedKey, value)
	}

	httpRequest, urlSuffix := e.resolveHttpRequest(request, payload)
	payload[fieldHttpRequest] = httpRequest
	payload[fieldURLSuffix] = urlSuffix

	if request.Name == "post_device" {
		if _, ok := payload["hostName"]; !ok {
			payload["hostName"] = e.hostName
		}
		if _, ok := payload["operationSystemName"]; !ok {
			payload["operationSystemName"] = e.operationSysName
		}
	}
	if _, ok := payload["updateDate"]; !ok && !request.Timestamp.IsZero() {
		payload["updateDate"] = request.Timestamp.UTC().Format(time.RFC3339)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal rabbitmq payload: %w", err)
	}

	e.logf("[info, rabbitmq enqueuer] publishing name=%s method=%s urlSuffix=%s", request.Name, httpRequest, urlSuffix)

	ctx, cancel := context.WithTimeout(context.Background(), e.publishTimeout)
	defer cancel()
	if err := e.publisher.Publish(ctx, body); err != nil {
		return fmt.Errorf("publish request %q: %w", request.Name, err)
	}

	return nil
}

func (e *RabbitMqEnqueuer) Close() error {
	if e == nil || e.publisher == nil {
		return nil
	}
	return e.publisher.Close()
}

func (e *RabbitMqEnqueuer) resolveHttpRequest(request Request, payload map[string]any) (string, string) {
	httpRequest := strings.ToUpper(readStringField(payload, fieldHttpRequest, "http_request"))
	if httpRequest == "" {
		switch strings.ToLower(strings.TrimSpace(request.Name)) {
		case "put_volume_change":
			httpRequest = "PUT"
		default:
			httpRequest = "POST"
		}
	}

	urlSuffix := readStringField(payload, fieldURLSuffix, "url_suffix")
	if urlSuffix == "" && httpRequest == "PUT" {
		pnpID := readStringField(payload, "pnpId", "pnp_id")
		urlSuffix = fmt.Sprintf("/%s/%s", pnpID, e.hostName)
	}

	return httpRequest, urlSuffix
}

func readStringField(payload map[string]any, keys ...string) string {
	for _, key := range keys {
		if v, ok := payload[key]; ok {
			s, okString := v.(string)
			if okString {
				return strings.TrimSpace(s)
			}
		}
	}
	return ""
}

func normalizeValue(mappedKey string, value string) any {
	trimmed := strings.TrimSpace(value)
	switch mappedKey {
	case "renderVolume", "captureVolume", "volume":
		if n, err := strconv.Atoi(trimmed); err == nil {
			return n
		}
	}
	return value
}

func mapFieldName(in string) string {
	if mapped, ok := snakeToCamelField[in]; ok {
		return mapped
	}
	return in
}

func (e *RabbitMqEnqueuer) logf(format string, args ...interface{}) {
	if e == nil || e.logger == nil {
		return
	}
	e.logger.Printf(format, args...)
}
