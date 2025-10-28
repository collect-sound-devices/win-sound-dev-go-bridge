package app

import (
	"context"
	"github.com/eduarddanziger/win-sound-dev-go-bridge/pkg/appinfo"
	"log"
	"os"
	"strings"

	"github.com/eduarddanziger/sound-win-scanner/v4/pkg/soundlibwrap"
)

var SaaHandle soundlibwrap.Handle

func Run(ctx context.Context) error {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	prefix := "log event handler,"
	// Bridge C log messages to Go logger.
	soundlibwrap.SetLogHandler(func(level, content string) {
		switch strings.ToLower(level) {
		case "trace", "debug":
			logger.Printf("[%s debug] %s", prefix, content)
		case "info":
			logger.Printf("[%s info] %s", prefix, content)
		case "warn", "warning":
			logger.Printf("[%s warn] %s", prefix, content)
		case "error", "critical":
			logger.Printf("[%s error] %s", prefix, content)
		default:
			logger.Printf("[%s log] %s", prefix, content)
		}
	})

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
		logger.Printf("[initially print default render] name=%q pnpId=%q vol=%d", desc.Name, desc.PnpID, desc.RenderVolume)
	} else {
		logger.Printf("[initially print default render] error: %v", err)
	}
	if desc, err := soundlibwrap.GetDefaultCapture(SaaHandle); err == nil {
		logger.Printf("[initially print default capture] name=%q pnpId=%q vol=%d", desc.Name, desc.PnpID, desc.CaptureVolume)
	} else {
		logger.Printf("[initially print default capture] error: %v", err)
	}

	// Keep running until interrupted to receive async logs and change events.
	<-ctx.Done()
	logger.Println("shutting down...")
	return nil
}
