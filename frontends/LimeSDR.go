package frontends

import (
  "github.com/myriadrf/limedrv"
	"github.com/quan-to/slog"
	"github.com/racerxdl/radioserver/protocol"
)

var limeLog = slog.Scope("LimeSDR Frontend")

type LimeSDRFrontend struct {
	device *limedrv.LMSDevice
	cb     SamplesCallback

  info    *protocol.DeviceInfo
  config  *protocol.DeviceConfig
  running  bool
}

func CreateLimeSDRFrontend(state *protocol.DeviceState) Frontend {
  devices := limedrv.GetDevices()
	if len(devices) == 0 {
		limeLog.Fatal("No devices found.\n")
	}

	var device = limedrv.Open(devices[0])

	var f = &LimeSDRFrontend{
		device:  device,
		running: false,
    info:    state.Info,
    config:  state.Config,
  }

	f.device.
		SetCallback(func(samples []complex64, _ int, _ uint64) {
			if f.cb != nil {
				f.cb(samples)
			}
		})


  // Global
  f.device.SetSampleRate(float64(f.config.SampleRate), int(f.config.Oversample))

  // RX CH0
  rxCh0 := f.config.RXC

	f.device.
    RXChannels[0].
		Enable().
		SetLPF(float64(f.config.SampleRate)).
		EnableLPF().
		SetDigitalLPF(float64(f.config.SampleRate)).
		EnableDigitalLPF().
		SetAntennaByName(rxCh0.Antenna).
    SetGainNormalized(float64(rxCh0.NormalizedGain)).
    SetCenterFrequency(float64(rxCh0.CenterFrequency))

	return f
}

func FindLimeSuiteDevices(dl *protocol.DeviceList) {
  devices := limedrv.GetDevices()

  for _, d := range devices {
    if d.Module == "FT601" {
      dl.Devices = append(dl.Devices, &LimeSDRMiniDefault)
    }
  }
}

func (f *LimeSDRFrontend) GetDeviceInfo() protocol.DeviceInfo {
  return *f.info
}

func (f *LimeSDRFrontend) GetDeviceConfig() protocol.DeviceConfig {
  return *f.config
}

func (f *LimeSDRFrontend) SetDeviceConfig() protocol.DeviceConfig {
  return *f.config
}

func (f *LimeSDRFrontend) Start() {
	if !f.running {
		limeLog.Info("Starting")
		f.device.Start()
		f.running = true
	}
}

func (f *LimeSDRFrontend) Stop() {
	if f.running {
		limeLog.Info("Stopping")
		f.device.Stop()
		f.running = false
	}
}

func (f *LimeSDRFrontend) SetSamplesAvailableCallback(cb SamplesCallback) {
	f.cb = cb
}

func (f *LimeSDRFrontend) Init() bool {
	return true
}

func (f *LimeSDRFrontend) Destroy() {

}
