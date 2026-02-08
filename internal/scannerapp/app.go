package scannerapp

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/collect-sound-devices/win-sound-dev-go-bridge/internal/enqueuer"
	"github.com/collect-sound-devices/win-sound-dev-go-bridge/pkg/appinfo"

	"github.com/collect-sound-devices/sound-win-scanner/v4/pkg/soundlibwrap"
)

var SaaHandle soundlibwrap.Handle

var logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lmicroseconds)

const (
	eventDefaultRenderChanged  = "default_render_changed"
	eventDefaultRenderRemoved  = "default_render_removed"
	eventDefaultCaptureChanged = "default_capture_changed"
	eventDefaultCaptureRemoved = "default_capture_removed"
	eventRenderVolumeChanged   = "render_volume_changed"
	eventCaptureVolumeChanged  = "capture_volume_changed"

	flowRender  = "render"
	flowCapture = "capture"
)

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

func postDeviceToApi(enqueue func(string, map[string]string), eventType, flowType, name, pnpID string, renderVolume, captureVolume *int) {
	fields := map[string]string{
		"device_message_type": eventType,
		"update_date":         time.Now().UTC().Format(time.RFC3339),
	}
	if flowType != "" {
		fields["flow_type"] = flowType
	}
	if name != "" {
		fields["name"] = name
	}
	if pnpID != "" {
		fields["pnp_id"] = pnpID
	}
	if renderVolume != nil {
		fields["render_volume"] = strconv.Itoa(*renderVolume)
	}
	if captureVolume != nil {
		fields["capture_volume"] = strconv.Itoa(*captureVolume)
	}

	enqueue("post_device", fields)
}

func putVolumeChangeToApi(enqueue func(string, map[string]string), eventType, pnpID string, volume int) {
	fields := map[string]string{
		"device_message_type": eventType,
		"update_date":         time.Now().UTC().Format(time.RFC3339),
		"volume":              strconv.Itoa(volume),
	}
	if pnpID != "" {
		fields["pnp_id"] = pnpID
	}

	enqueue("put_volume_change", fields)
}

func Run(ctx context.Context) error {
	reqEnqueuer := enqueuer.NewEmptyRequestEnqueuer(logger)
	enqueue := func(name string, fields map[string]string) {
		if err := reqEnqueuer.EnqueueRequest(enqueuer.Request{
			Name:      name,
			Timestamp: time.Now(),
			Fields:    fields,
		}); err != nil {
			logError("enqueue failed: %v", err)
		}
	}

	{
		logHandlerLogger := log.New(os.Stdout, "", 0)
		prefix := "cpp backend,"
		// Bridge C soundlibwrap messages to Go logHandlerLogger.
		soundlibwrap.SetLogHandler(func(timestamp, level, content string) {
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
				renderVolume := int(desc.RenderVolume)
				postDeviceToApi(enqueue, eventDefaultRenderChanged, flowRender, desc.Name, desc.PnpID, &renderVolume, nil)
				logInfo("Render device changed: name=%q pnpId=%q vol=%d", desc.Name, desc.PnpID, desc.RenderVolume)
			} else {
				logError("Render device changed, can not read it: %v", err)
			}
		} else {
			postDeviceToApi(enqueue, eventDefaultRenderRemoved, flowRender, "", "", nil, nil)
			logInfo("Render device removed")
		}

	})
	soundlibwrap.SetDefaultCaptureHandler(func(present bool) {
		if present {
			if desc, err := soundlibwrap.GetDefaultCapture(SaaHandle); err == nil {
				captureVolume := int(desc.CaptureVolume)
				postDeviceToApi(enqueue, eventDefaultCaptureChanged, flowCapture, desc.Name, desc.PnpID, nil, &captureVolume)
				logInfo("Capture device changed: name=%q pnpId=%q vol=%d", desc.Name, desc.PnpID, desc.RenderVolume)
			} else {
				logError("Capture device changed, can not read it: %v", err)
			}
		} else {
			postDeviceToApi(enqueue, eventDefaultCaptureRemoved, flowCapture, "", "", nil, nil)
			logInfo("Capture device removed")
		}
	})

	// Volume change notifications.
	soundlibwrap.SetRenderVolumeChangedHandler(func() {
		if desc, err := soundlibwrap.GetDefaultRender(SaaHandle); err == nil {
			putVolumeChangeToApi(enqueue, eventRenderVolumeChanged, desc.PnpID, int(desc.RenderVolume))
			logInfo("Render volume changed: name=%q pnpId=%q vol=%d", desc.Name, desc.PnpID, desc.RenderVolume)
		} else {
			logError("Render volume changed, can not read it: %v", err)
		}
	})
	soundlibwrap.SetCaptureVolumeChangedHandler(func() {
		if desc, err := soundlibwrap.GetDefaultCapture(SaaHandle); err == nil {
			putVolumeChangeToApi(enqueue, eventCaptureVolumeChanged, desc.PnpID, int(desc.CaptureVolume))
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
