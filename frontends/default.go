package frontends

import (
  "github.com/racerxdl/radioserver/protocol"
)

// LimeSDR Mini
var LimeSDRMiniDefault = protocol.DeviceInfo{
  Name: 3,
  MaximumSampleRate: 30.72e6,
  MinimumFrequency: 10e6,
  MaximumFrequency: 3.5e9,
  ADCResolution: 12,
  MaximumRXChannels: 1,
  MaximumTXChannels: 1,
}
