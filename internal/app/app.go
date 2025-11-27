package app

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/eduarddanziger/win-sound-dev-go-bridge/pkg/appinfo"

	"github.com/eduarddanziger/sound-win-scanner/v4/pkg/soundlibwrap"
)

var SaaHandle soundlibwrap.Handle

var logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lmicroseconds)

func Run(ctx context.Context) error {

	{
		logHandlerLogger := log.New(os.Stdout, "", 0)
		prefix := "Log event handler,"
		// Bridge C logHandlerLogger messages to Go logHandlerLogger.
		soundlibwrap.SetLogHandler(func(ts, level, content string) {
			// Prefix each logHandlerLogger from the C side with a timestamp (microseconds)
			switch strings.ToLower(level) {
			case "trace", "debug":
				logHandlerLogger.Printf("%s [%s debug] %s", ts, prefix, content)
			case "info":
				logHandlerLogger.Printf("%s [%s info] %s", ts, prefix, content)
			case "warn", "warning":
				logHandlerLogger.Printf("%s [%s warn] %s", ts, prefix, content)
			case "error", "critical":
				logHandlerLogger.Printf("%s [%s error] %s", ts, prefix, content)
			default:
				logHandlerLogger.Printf("%s [%s info] %s", ts, prefix, content)
			}
		})
	}

	// Device default change notifications.
	soundlibwrap.SetDefaultRenderHandler(func(present bool) {
		logger.Printf("[default render handler] present=%v", present)
		if present {
			if desc, err := soundlibwrap.GetDefaultRender(SaaHandle); err == nil {
				logger.Printf("[default render handler] name=%q pnpId=%q vol=%d", desc.Name, desc.PnpID, desc.RenderVolume)
			} else {
				logger.Printf("[default render handler] error: %v", err)
			}
		}
	})
	soundlibwrap.SetDefaultCaptureHandler(func(present bool) {
		logger.Printf("[default capture handler] present=%v", present)
		if present {
			if desc, err := soundlibwrap.GetDefaultCapture(SaaHandle); err == nil {
				logger.Printf("[default capture handler] name=%q pnpId=%q vol=%d", desc.Name, desc.PnpID, desc.RenderVolume)
			} else {
				logger.Printf("[default capture handler] error: %v", err)
			}
		}
	})

	logger.Println("Initializing...")

	// Initialize the C library and register callbacks using the global handle.
	var err error
	SaaHandle, err = soundlibwrap.Initialize(appinfo.AppName, appinfo.Version)
	if err != nil {
		return err
	}
	defer func() {
		_ = soundlibwrap.Uninitialize(SaaHandle)
		SaaHandle = 0
	}()

	if err := soundlibwrap.RegisterCallbacks(SaaHandle); err != nil {
		return err
	}

	// Print the default render and capture devices.
	if desc, err := soundlibwrap.GetDefaultRender(SaaHandle); err == nil {
		logger.Printf("Default render device: name=%q pnpId=%q vol=%d", desc.Name, desc.PnpID, desc.RenderVolume)
	} else {
		logger.Printf("Default render device error: %v", err)
	}
	if desc, err := soundlibwrap.GetDefaultCapture(SaaHandle); err == nil {
		logger.Printf("Default capture device: name=%q pnpId=%q vol=%d", desc.Name, desc.PnpID, desc.CaptureVolume)
	} else {
		logger.Printf("Default capture device error: %v", err)
	}

	// Keep running until interrupted to receive async logs and change events.
	<-ctx.Done()
	logger.Println("shutting down...")
	return nil
}
