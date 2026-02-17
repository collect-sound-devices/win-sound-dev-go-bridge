package contract

type MessageType uint8

const (
	MessageTypeVolumeRenderChanged   MessageType = 3
	MessageTypeVolumeCaptureChanged  MessageType = 4
	MessageTypeDefaultRenderChanged  MessageType = 5
	MessageTypeDefaultCaptureChanged MessageType = 6
)

const (
	RequestPostDevice      = "post_device"
	RequestPutVolumeChange = "put_volume_change"
)

type FlowType uint8

const (
	FlowTypeRender  FlowType = 1
	FlowTypeCapture FlowType = 2
)

const (
	FieldDeviceMessageType   = "deviceMessageType"
	FieldUpdateDate          = "updateDate"
	FieldFlowType            = "flowType"
	FieldName                = "name"
	FieldPnpID               = "pnpId"
	FieldRenderVolume        = "renderVolume"
	FieldCaptureVolume       = "captureVolume"
	FieldVolume              = "volume"
	FieldHostName            = "hostName"
	FieldOperationSystemName = "operationSystemName"
	FieldHTTPRequest         = "httpRequest"
	FieldURLSuffix           = "urlSuffix"
)
