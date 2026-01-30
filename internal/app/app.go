package app

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/collect-sound-devices/win-sound-dev-go-bridge/pkg/appinfo"

	"github.com/collect-sound-devices/sound-win-scanner/v4/pkg/soundlibwrap"
)

var SaaHandle soundlibwrap.Handle

var logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lmicroseconds)

func logf(level, format string, v ...interface{}) {
	if level == "" {
		level = "info"
	}
	logger.Printf("["+level+"] "+format, v...)
}

func logInfo(format string, v ...interface{}) {
	logf("info", format, v...)
}

func logError(format string, v ...interface{}) {
	logf("error", format, v...)
}

func Run(ctx context.Context) error {

	{
		logHandlerLogger := log.New(os.Stdout, "", 0)
		prefix := "cpp backend,"
		// Bridge C logHandlerLogger messages to Go logHandlerLogger.
		soundlibwrap.SetLogHandler(func(timestamp, level, content string) {
			// Prefix each logHandlerLogger from the C side with a timestamp (microseconds)
			switch strings.ToLower(level) {
			case "trace", "debug":
				logHandlerLogger.Printf("%s [%s debug] %s", timestamp, prefix, content)
			case "info":
				logHandlerLogger.Printf("%s [%s info] %s", timestamp, prefix, content)
			case "warn", "warning":
				logHandlerLogger.Printf("%s [%s warn] %s", timestamp, prefix, content)
			case "error", "critical":
				logHandlerLogger.Printf("%s [%s error] %s", timestamp, prefix, content)
			default:
				logHandlerLogger.Printf("%s [%s info] %s", timestamp, prefix, content)
			}
		})
	}

	// Device default change notifications.
	soundlibwrap.SetDefaultRenderHandler(func(present bool) {
		if present {
			if desc, err := soundlibwrap.GetDefaultRender(SaaHandle); err == nil {
				logInfo("Render device changed: name=%q pnpId=%q vol=%d", desc.Name, desc.PnpID, desc.RenderVolume)
			} else {
				logError("Render device changed, can not read it: %v", err)
			}
		} else {
			logInfo("Render device removed")
		}

	})
	soundlibwrap.SetDefaultCaptureHandler(func(present bool) {
		if present {
			if desc, err := soundlibwrap.GetDefaultCapture(SaaHandle); err == nil {
				logInfo("Capture device changed: name=%q pnpId=%q vol=%d", desc.Name, desc.PnpID, desc.RenderVolume)
			} else {
				logError("Capture device changed, can not read it: %v", err)
			}
		} else {
			logInfo("Capture device removed")
		}
	})

	// Volume change notifications.
	soundlibwrap.SetRenderVolumeChangedHandler(func() {
		if desc, err := soundlibwrap.GetDefaultRender(SaaHandle); err == nil {
			logInfo("Render volume changed: name=%q pnpId=%q vol=%d", desc.Name, desc.PnpID, desc.RenderVolume)
		} else {
			logError("Render volume changed, can not read it: %v", err)
		}
	})
	soundlibwrap.SetCaptureVolumeChangedHandler(func() {
		if desc, err := soundlibwrap.GetDefaultCapture(SaaHandle); err == nil {
			logInfo("Capture volume changed: name=%q pnpId=%q vol=%d", desc.Name, desc.PnpID, desc.CaptureVolume)
		} else {
			logError("Capture volume changed, can not read it: %v", err)
		}
	})

	logInfo("Initializing...")

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
		if desc.PnpID == "" {
			logInfo("No default render device.")
		} else {
			logInfo("Render device info: name=%q pnpId=%q vol=%d", desc.Name, desc.PnpID, desc.RenderVolume)
		}
	} else {
		logError("Render device info, can not read it: %v", err)
	}
	if desc, err := soundlibwrap.GetDefaultCapture(SaaHandle); err == nil {
		if desc.PnpID == "" {
			logInfo("No default capture device.")
		} else {
			logInfo("Capture device info: name=%q pnpId=%q vol=%d", desc.Name, desc.PnpID, desc.RenderVolume)
		}
	} else {
		logError("Capture device info, can not read it: %v", err)
	}

	// Keep running until interrupted to receive async logs and change events.
	<-ctx.Done()
	logInfo("Shutting down...")
	return nil
}
