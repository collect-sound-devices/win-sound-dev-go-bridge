package scannerapp

import (
	"strconv"
	"time"

	"github.com/collect-sound-devices/sound-win-scanner/v4/pkg/soundlibwrap"
	c "github.com/collect-sound-devices/win-sound-dev-go-bridge/internal/contract"
	"github.com/collect-sound-devices/win-sound-dev-go-bridge/pkg/appinfo"
)

type ScannerApp interface {
	RepostRenderDeviceToApi(c.EventType)
	RepostCaptureDeviceToApi(c.EventType)
	Shutdown()
}

type scannerAppImpl struct {
	soundLibHandle soundlibwrap.Handle
	enqueueFunc    func(c.EventType, map[string]string)
	logInfo        func(string, ...interface{})
	logError       func(string, ...interface{})
}

func NewImpl(enqueue func(c.EventType, map[string]string), logInfo func(string, ...interface{}), logError func(string, ...interface{})) (*scannerAppImpl, error) {
	app := &scannerAppImpl{
		enqueueFunc: enqueue,
		logInfo:     logInfo,
		logError:    logError,
	}
	app.attachHandlers()
	if err := app.init(); err != nil {
		return nil, err
	}

	// Post the default render and capture devices.
	app.RepostRenderDeviceToApi(c.EventTypeRenderDeviceConfirmed)
	app.RepostCaptureDeviceToApi(c.EventTypeCaptureDeviceConfirmed)

	return app, nil
}

func (app *scannerAppImpl) init() error {
	h, err := soundlibwrap.Initialize(appinfo.AppName, appinfo.Version)
	if err != nil {
		return err
	}
	app.soundLibHandle = h
	if err := soundlibwrap.RegisterCallbacks(app.soundLibHandle); err != nil {
		_ = soundlibwrap.Uninitialize(app.soundLibHandle)
		app.soundLibHandle = 0
		return err
	}
	return nil
}

func (app *scannerAppImpl) attachHandlers() {
	// Device default change notifications.
	soundlibwrap.SetDefaultRenderHandler(func(present bool) {
		if present {
			app.RepostRenderDeviceToApi(c.EventTypeRenderDeviceDiscovered)
		} else {
			// not yet implemented removeDeviceToApi
			app.logInfo("Render device removed")
		}
	})
	soundlibwrap.SetDefaultCaptureHandler(func(present bool) {
		if present {
			app.RepostCaptureDeviceToApi(c.EventTypeCaptureDeviceDiscovered)
		} else {
			// not yet implemented removeDeviceToApi
			app.logInfo("Capture device removed")
		}
	})

	// Volume change notifications.
	soundlibwrap.SetRenderVolumeChangedHandler(func() {
		if desc, err := soundlibwrap.GetDefaultRender(app.soundLibHandle); err == nil {
			app.putVolumeChangeToApi(c.EventTypeRenderVolumeChanged, desc.PnpID, int(desc.RenderVolume))
			app.logInfo("Render volume changed: name=%q pnpId=%q vol=%d", desc.Name, desc.PnpID, desc.RenderVolume)
		} else {
			app.logError("Render volume changed, can not read it: %v", err)
		}
	})
	soundlibwrap.SetCaptureVolumeChangedHandler(func() {
		if desc, err := soundlibwrap.GetDefaultCapture(app.soundLibHandle); err == nil {
			app.putVolumeChangeToApi(c.EventTypeCaptureVolumeChanged, desc.PnpID, int(desc.CaptureVolume))
			app.logInfo("Capture volume changed: name=%q pnpId=%q vol=%d", desc.Name, desc.PnpID, desc.CaptureVolume)
		} else {
			app.logError("Capture volume changed, can not read it: %v", err)
		}
	})
}

func (app *scannerAppImpl) Shutdown() {
	if app.soundLibHandle != 0 {
		_ = soundlibwrap.Uninitialize(app.soundLibHandle)
		app.soundLibHandle = 0
	}
}

func (app *scannerAppImpl) putVolumeChangeToApi(event c.EventType, pnpID string, volume int) {
	fields := map[string]string{
		c.FieldUpdateDate: time.Now().UTC().Format(time.RFC3339),
		c.FieldVolume:     strconv.Itoa(volume),
	}
	if pnpID != "" {
		fields[c.FieldPnpID] = pnpID
	}

	app.enqueueFunc(event, fields)
}

func (app *scannerAppImpl) RepostRenderDeviceToApi(event c.EventType) {
	if desc, err := soundlibwrap.GetDefaultRender(app.soundLibHandle); err == nil {
		renderVolume := int(desc.RenderVolume)
		captureVolume := int(desc.CaptureVolume)
		app.postDeviceToApi(event, desc.Name, desc.PnpID, renderVolume, captureVolume)
		app.logInfo("Render device identified and updated: name=%q pnpId=%q renderVol=%d captureVol=%d", desc.Name, desc.PnpID, desc.RenderVolume, desc.CaptureVolume)
	} else {
		app.logError("Render device can not be identified: %v", err)
	}
}

func (app *scannerAppImpl) RepostCaptureDeviceToApi(event c.EventType) {
	if desc, err := soundlibwrap.GetDefaultCapture(app.soundLibHandle); err == nil {
		renderVolume := int(desc.RenderVolume)
		captureVolume := int(desc.CaptureVolume)
		app.postDeviceToApi(event, desc.Name, desc.PnpID, renderVolume, captureVolume)
		app.logInfo("Capture device identified and updated: name=%q pnpId=%q renderVol=%d captureVol=%d", desc.Name, desc.PnpID, desc.RenderVolume, desc.CaptureVolume)
	} else {
		app.logError("Capture device can not be identified: %v", err)
	}
}

func (app *scannerAppImpl) postDeviceToApi(event c.EventType, name, pnpID string, renderVolume, captureVolume int) {
	fields := map[string]string{
		c.FieldUpdateDate:    time.Now().UTC().Format(time.RFC3339),
		c.FieldName:          name,
		c.FieldPnpID:         pnpID,
		c.FieldRenderVolume:  strconv.Itoa(renderVolume),
		c.FieldCaptureVolume: strconv.Itoa(captureVolume),
	}

	app.enqueueFunc(event, fields)
}
