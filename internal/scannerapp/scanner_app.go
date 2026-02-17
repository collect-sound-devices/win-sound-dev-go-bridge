package scannerapp

import (
	"strconv"
	"time"

	"github.com/collect-sound-devices/win-sound-dev-go-bridge/internal/contract"
	"github.com/collect-sound-devices/win-sound-dev-go-bridge/pkg/appinfo"

	"github.com/collect-sound-devices/sound-win-scanner/v4/pkg/soundlibwrap"
)

type ScannerApp interface {
	RepostRenderDeviceToApi()
	RepostCaptureDeviceToApi()
	Shutdown()
}

type scannerAppImpl struct {
	soundLibHandle soundlibwrap.Handle
	enqueueFunc    func(uint8, map[string]string)
	logInfo        func(string, ...interface{})
	logError       func(string, ...interface{})
}

func NewImpl(enqueue func(uint8, map[string]string), logInfo func(string, ...interface{}), logError func(string, ...interface{})) (*scannerAppImpl, error) {
	a := &scannerAppImpl{
		enqueueFunc: enqueue,
		logInfo:     logInfo,
		logError:    logError,
	}
	a.attachHandlers()
	if err := a.init(); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *scannerAppImpl) init() error {
	h, err := soundlibwrap.Initialize(appinfo.AppName, appinfo.Version)
	if err != nil {
		return err
	}
	a.soundLibHandle = h
	if err := soundlibwrap.RegisterCallbacks(a.soundLibHandle); err != nil {
		_ = soundlibwrap.Uninitialize(a.soundLibHandle)
		a.soundLibHandle = 0
		return err
	}
	return nil
}

func (a *scannerAppImpl) attachHandlers() {
	// Device default change notifications.
	soundlibwrap.SetDefaultRenderHandler(func(present bool) {
		if present {
			a.RepostRenderDeviceToApi()
		} else {
			// not yet implemented removeDeviceToApi
			a.logInfo("Render device removed")
		}
	})
	soundlibwrap.SetDefaultCaptureHandler(func(present bool) {
		if present {
			a.RepostCaptureDeviceToApi()
		} else {
			// not yet implemented removeDeviceToApi
			a.logInfo("Capture device removed")
		}
	})

	// Volume change notifications.
	soundlibwrap.SetRenderVolumeChangedHandler(func() {
		if desc, err := soundlibwrap.GetDefaultRender(a.soundLibHandle); err == nil {
			a.putVolumeChangeToApi(uint8(contract.MessageTypeVolumeRenderChanged), desc.PnpID, int(desc.RenderVolume))
			a.logInfo("Render volume changed: name=%q pnpId=%q vol=%d", desc.Name, desc.PnpID, desc.RenderVolume)
		} else {
			a.logError("Render volume changed, can not read it: %v", err)
		}
	})
	soundlibwrap.SetCaptureVolumeChangedHandler(func() {
		if desc, err := soundlibwrap.GetDefaultCapture(a.soundLibHandle); err == nil {
			a.putVolumeChangeToApi(uint8(contract.MessageTypeVolumeCaptureChanged), desc.PnpID, int(desc.CaptureVolume))
			a.logInfo("Capture volume changed: name=%q pnpId=%q vol=%d", desc.Name, desc.PnpID, desc.CaptureVolume)
		} else {
			a.logError("Capture volume changed, can not read it: %v", err)
		}
	})
}

func (a *scannerAppImpl) Shutdown() {
	if a.soundLibHandle != 0 {
		_ = soundlibwrap.Uninitialize(a.soundLibHandle)
		a.soundLibHandle = 0
	}
}

func (a *scannerAppImpl) putVolumeChangeToApi(messageType uint8, pnpID string, volume int) {
	fields := map[string]string{
		contract.FieldUpdateDate: time.Now().UTC().Format(time.RFC3339),
		contract.FieldVolume:     strconv.Itoa(volume),
	}
	if pnpID != "" {
		fields[contract.FieldPnpID] = pnpID
	}

	a.enqueueFunc(messageType, fields)
}

func (a *scannerAppImpl) RepostRenderDeviceToApi() {
	if desc, err := soundlibwrap.GetDefaultRender(a.soundLibHandle); err == nil {
		renderVolume := int(desc.RenderVolume)
		captureVolume := int(desc.CaptureVolume)
		a.postDeviceToApi(uint8(contract.MessageTypeDefaultRenderChanged), desc.Name, desc.PnpID, renderVolume, captureVolume)
		a.logInfo("Render device identified and updated: name=%q pnpId=%q renderVol=%d captureVol=%d", desc.Name, desc.PnpID, desc.RenderVolume, desc.CaptureVolume)
	} else {
		a.logError("Render device can not be identified: %v", err)
	}
}

func (a *scannerAppImpl) RepostCaptureDeviceToApi() {
	if desc, err := soundlibwrap.GetDefaultCapture(a.soundLibHandle); err == nil {
		renderVolume := int(desc.RenderVolume)
		captureVolume := int(desc.CaptureVolume)
		a.postDeviceToApi(uint8(contract.MessageTypeDefaultCaptureChanged), desc.Name, desc.PnpID, renderVolume, captureVolume)
		a.logInfo("Capture device identified and updated: name=%q pnpId=%q renderVol=%d captureVol=%d", desc.Name, desc.PnpID, desc.RenderVolume, desc.CaptureVolume)
	} else {
		a.logError("Capture device can not be identified: %v", err)
	}
}

func (a *scannerAppImpl) postDeviceToApi(messageType uint8, name, pnpID string, renderVolume, captureVolume int) {
	fields := map[string]string{
		contract.FieldUpdateDate:    time.Now().UTC().Format(time.RFC3339),
		contract.FieldName:          name,
		contract.FieldPnpID:         pnpID,
		contract.FieldRenderVolume:  strconv.Itoa(renderVolume),
		contract.FieldCaptureVolume: strconv.Itoa(captureVolume),
	}

	a.enqueueFunc(messageType, fields)
}
