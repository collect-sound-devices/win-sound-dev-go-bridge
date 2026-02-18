package scannerapp

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	c "github.com/collect-sound-devices/win-sound-dev-go-bridge/internal/contract"
	"github.com/collect-sound-devices/win-sound-dev-go-bridge/internal/enqueuer"
	"github.com/collect-sound-devices/win-sound-dev-go-bridge/internal/logging"
	"github.com/collect-sound-devices/win-sound-dev-go-bridge/internal/rabbitmq"
)

func NewWithLogger(enqueue func(c.EventType, map[string]string), logger logging.Logger) (ScannerApp, error) {
	return NewImpl(
		enqueue,
		func(format string, v ...interface{}) { logging.PrintInfo(logger, format, v...) },
		func(format string, v ...interface{}) { logging.PrintError(logger, format, v...) },
	)
}

func Run(ctx context.Context) error {
	appLogger := logging.NewAppLogger()
	reqEnqueuer, cleanupEnqueuer, err := newRequestEnqueuer(ctx, appLogger)
	if err != nil {
		return err
	}
	defer cleanupEnqueuer()

	enqueue := func(event c.EventType, fields map[string]string) {
		if err := reqEnqueuer.EnqueueRequest(enqueuer.Request{
			Timestamp: time.Now(),
			Event:     event,
			Fields:    fields,
		}); err != nil {
			logging.PrintError(appLogger, "enqueue failed: %v", err)
		}
	}

	{
		logging.AttachSoundlibwrapBridge(logging.NewPlainLogger(), "cpp backend,")
	}

	logging.PrintInfo(appLogger, "Initializing...")

	app, err := NewWithLogger(enqueue, appLogger)
	if err != nil {
		return err
	}
	defer app.Shutdown()

	// Keep running until interrupted to receive async logs and change events.
	<-ctx.Done()
	logging.PrintInfo(appLogger, "Shutting down...")
	return nil
}

func newRequestEnqueuer(ctx context.Context, logger logging.Logger) (enqueuer.EnqueueRequest, func(), error) {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("WIN_SOUND_ENQUEUER")))

	// Return a no-op enqueuer for testing or when RabbitMQ is not available.
	if mode == "empty" {
		return enqueuer.NewEmptyRequestEnqueuer(logger), func() {}, nil
	}

	// Validate that the configured mode is supported.
	if mode != "" && mode != "rabbitmq" {
		return nil, nil, fmt.Errorf("unsupported WIN_SOUND_ENQUEUER=%q (supported: empty, rabbitmq)", mode)
	}

	cfg, err := rabbitmq.LoadConfigFromEnv()
	if err != nil {
		return nil, nil, err
	}

	publisher, err := rabbitmq.NewRequestPublisher(ctx, cfg, logger)
	if err != nil {
		return nil, nil, err
	}

	reqEnqueuer := rabbitmq.NewRabbitMqEnqueuerWithContext(ctx, publisher, logger)
	cleanup := func() {
		if err := reqEnqueuer.Close(); err != nil {
			logging.PrintError(logger, "rabbitmq enqueuer close failed: %v", err)
		}
	}

	return reqEnqueuer, cleanup, nil
}
