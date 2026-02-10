package scannerapp

import (
	"context"
	"time"

	"github.com/collect-sound-devices/win-sound-dev-go-bridge/internal/enqueuer"
	"github.com/collect-sound-devices/win-sound-dev-go-bridge/internal/logging"
)

func NewWithLogger(enqueue func(string, map[string]string), logger logging.Logger) (ScannerApp, error) {
	return NewImpl(
		enqueue,
		func(format string, v ...interface{}) { logging.PrintInfo(logger, format, v...) },
		func(format string, v ...interface{}) { logging.PrintError(logger, format, v...) },
	)
}

func Run(ctx context.Context) error {
	appLogger := logging.NewAppLogger()
	reqEnqueuer := enqueuer.NewEmptyRequestEnqueuer(appLogger)
	enqueue := func(name string, fields map[string]string) {
		if err := reqEnqueuer.EnqueueRequest(enqueuer.Request{
			Name:      name,
			Timestamp: time.Now(),
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

	// Post the default render and capture devices.
	app.RepostRenderDeviceToApi()
	app.RepostCaptureDeviceToApi()

	// Keep running until interrupted to receive async logs and change events.
	<-ctx.Done()
	logging.PrintInfo(appLogger, "Shutting down...")
	return nil
}
