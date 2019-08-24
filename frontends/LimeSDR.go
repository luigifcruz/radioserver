package frontends

import (
	"strconv"

	"github.com/luigifreitas/radioserver/protocol"
	"github.com/myriadrf/limedrv"
	"github.com/quan-to/slog"
)

var limeLog = slog.Scope("LimeSDR Frontend")

type LimeSDRFrontend struct {
	device *limedrv.LMSDevice
	cb     SamplesCallback

	info    *protocol.DeviceInfo
	config  *protocol.DeviceConfig
	running bool
}

func CreateLimeSDRFrontend(state *protocol.DeviceState) Frontend {
	devices := limedrv.GetDevices()

	i, _ := strconv.Atoi(state.Info.Serial)
	var device = limedrv.Open(devices[i])

	var f = &LimeSDRFrontend{
		device:  device,
		running: false,
		info:    state.Info,
		config:  &protocol.DeviceConfig{},
	}

	f.device.
		SetCallback(func(samples []complex64, _ int, _ uint64) {
			if f.cb != nil {
				f.cb(samples)
			}
		})

	f.device.SetSampleRate(float64(state.Config.SampleRate), int(state.Config.Oversample))
	f.SetDeviceConfig(*state.Config)

	return f
}

func FindLimeSuiteDevices(dl *protocol.DeviceList) {
	devices := limedrv.GetDevices()

	for i, d := range devices {
		if d.Module == "FT601" {
			b := LimeSDRMiniDefault
			b.Serial = strconv.Itoa(i)
			dl.Devices = append(dl.Devices, &b)
		}
	}
}

func (f *LimeSDRFrontend) GetDeviceInfo() protocol.DeviceInfo {
	return *f.info
}

func (f *LimeSDRFrontend) GetDeviceConfig() protocol.DeviceConfig {
	return *f.config
}

func (f *LimeSDRFrontend) SetDeviceConfig(c protocol.DeviceConfig) protocol.DeviceConfig {

	for i, n := range c.RXC {
		o := &protocol.ChannelConfig{}
		if len(f.config.RXC) > i {
			o = f.config.RXC[i]
		}

		f.device.RXChannels[i].
      Enable()
//      SetLPF(float64(5e6)).
//      EnableLPF().
//      SetDigitalLPF(float64(5e6)).
//      EnableDigitalLPF()

		if n.NormalizedGain != o.NormalizedGain {
			f.device.SetGainNormalized(i, true, float64(n.NormalizedGain))
		}

		if n.Antenna != o.Antenna {
			f.device.SetAntennaByName(n.Antenna, i, true)
		}

		if n.CenterFrequency != o.CenterFrequency {
			f.device.SetCenterFrequency(i, true, float64(n.CenterFrequency))
			limeLog.Info("Channel %d: Tuning center frequency: %v", i, n.CenterFrequency)
		}
	}

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
