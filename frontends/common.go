package frontends

import (
	"github.com/luigifreitas/radioserver/protocol"
)

const (
	SampleTypeFloatIQ = iota
	SampleTypeS16IQ
	SampleTypeS8IQ
)

const minimumSampleRate = 10e3

type Frontend interface {
	GetDeviceInfo() protocol.DeviceInfo
	GetDeviceConfig() protocol.DeviceConfig
	SetDeviceConfig(*protocol.DeviceConfig) protocol.DeviceConfig

	Init() bool
	Start()
	Stop()
	Destroy()
	SetSamplesAvailableCallback(cb SamplesCallback)
}

type SamplesCallback func(samples []complex64)

type Frontends map[string]func(*protocol.DeviceState) Frontend
type Find map[string]func(*protocol.DeviceList)

var FindDevices = Find{
	"LimeSuite": FindLimeSuiteDevices,
}

var Available = Frontends{
	"LimeSDRMini": CreateLimeSDRFrontend,
	//  "AirspyMini": CreateAirspyFrontend,
}
