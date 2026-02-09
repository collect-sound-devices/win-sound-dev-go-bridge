package scannerapp

import (
	"strconv"
	"time"

	"github.com/collect-sound-devices/win-sound-dev-go-bridge/pkg/appinfo"

	"github.com/collect-sound-devices/sound-win-scanner/v4/pkg/soundlibwrap"
)

type ScannerApp struct {
	soundLibHandle soundlibwrap.Handle
	enqueueFunc    func(string, map[string]string)
}

func (a *ScannerApp) init() error {
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

func (a *ScannerApp) shutdown() {
	if a.soundLibHandle != 0 {
		_ = soundlibwrap.Uninitialize(a.soundLibHandle)
		a.soundLibHandle = 0
	}
}

func (a *ScannerApp) postDeviceToApi(eventType, flowType, name, pnpID string, renderVolume, captureVolume int) {
	fields := map[string]string{
		"device_message_type": eventType,
		"update_date":         time.Now().UTC().Format(time.RFC3339),
		"flow_type":           flowType,
		"name":                name,
		"pnp_id":              pnpID,
		"render_volume":       strconv.Itoa(renderVolume),
		"capture_volume":      strconv.Itoa(captureVolume),
	}

	a.enqueueFunc("post_device", fields)
}

func (a *ScannerApp) putVolumeChangeToApi(eventType, pnpID string, volume int) {
	fields := map[string]string{
		"device_message_type": eventType,
		"update_date":         time.Now().UTC().Format(time.RFC3339),
		"volume":              strconv.Itoa(volume),
	}
	if pnpID != "" {
		fields["pnp_id"] = pnpID
	}

	a.enqueueFunc("put_volume_change", fields)
}

func (a *ScannerApp) attachHandlers() {
	// Device default change notifications.
	soundlibwrap.SetDefaultRenderHandler(func(present bool) {
		if present {
			if desc, err := soundlibwrap.GetDefaultRender(a.soundLibHandle); err == nil {
				renderVolume := int(desc.RenderVolume)
				captureVolume := int(desc.CaptureVolume)
				a.postDeviceToApi(eventDefaultRenderChanged, flowRender, desc.Name, desc.PnpID, renderVolume, captureVolume)
				logInfo("Render device changed: name=%q pnpId=%q renderVol=%d captureVol=%d", desc.Name, desc.PnpID, desc.RenderVolume, desc.CaptureVolume)
			} else {
				logError("Render device changed, can not read it: %v", err)
			}
		} else {
			// not yet implemented removeDeviceToApi
			logInfo("Render device removed")
		}

	})
	soundlibwrap.SetDefaultCaptureHandler(func(present bool) {
		if present {
			if desc, err := soundlibwrap.GetDefaultCapture(a.soundLibHandle); err == nil {
				renderVolume := int(desc.RenderVolume)
				captureVolume := int(desc.CaptureVolume)
				a.postDeviceToApi(eventDefaultCaptureChanged, flowCapture, desc.Name, desc.PnpID, renderVolume, captureVolume)
				logInfo("Capture device changed: name=%q pnpId=%q renderVol=%d captureVol=%d", desc.Name, desc.PnpID, desc.RenderVolume, desc.CaptureVolume)
			} else {
				logError("Capture device changed, can not read it: %v", err)
			}
		} else {
			// not yet implemented removeDeviceToApi
			logInfo("Capture device removed")
		}
	})

	// Volume change notifications.
	soundlibwrap.SetRenderVolumeChangedHandler(func() {
		if desc, err := soundlibwrap.GetDefaultRender(a.soundLibHandle); err == nil {
			a.putVolumeChangeToApi(eventRenderVolumeChanged, desc.PnpID, int(desc.RenderVolume))
			logInfo("Render volume changed: name=%q pnpId=%q vol=%d", desc.Name, desc.PnpID, desc.RenderVolume)
		} else {
			logError("Render volume changed, can not read it: %v", err)
		}
	})
	soundlibwrap.SetCaptureVolumeChangedHandler(func() {
		if desc, err := soundlibwrap.GetDefaultCapture(a.soundLibHandle); err == nil {
			a.putVolumeChangeToApi(eventCaptureVolumeChanged, desc.PnpID, int(desc.CaptureVolume))
			logInfo("Capture volume changed: name=%q pnpId=%q vol=%d", desc.Name, desc.PnpID, desc.CaptureVolume)
		} else {
			logError("Capture volume changed, can not read it: %v", err)
		}
	})
}
