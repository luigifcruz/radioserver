package client

import (
	"context"
	"github.com/quan-to/slog"
	"github.com/racerxdl/radioserver/protocol"
	"google.golang.org/grpc"
)

var log = slog.Scope("RadioClient")

type Callback interface {
	OnData([]complex64)
}

type RadioClient struct {
	name           string
	app            string
	address        string
	routineRunning bool
	terminated     bool
	conn           *grpc.ClientConn
	client         protocol.RadioServerClient
	serverInfo     *protocol.ServerInfoData
	deviceInfo     *protocol.DeviceInfo

	currentSampleRate      uint32
	availableSampleRates   []uint32

	iqChannelConfig      *protocol.ChannelConfig
	iqChannelEnabled      bool

	gain      uint32
	streaming bool
	cb        Callback
}

func MakeRadioClient(address, name, application string) *RadioClient {
	return &RadioClient{
		name:                  name,
		app:                   application,
		address:               address,
		routineRunning:        false,
		availableSampleRates:  []uint32{},
		iqChannelConfig:       &protocol.ChannelConfig{},
		iqChannelEnabled:      false,
		streaming:             false,
    currentSampleRate: 3000000,
  }
}

// region Public Methods

// GetName returns the name of the active device in RadioClient
func (f *RadioClient) GetName() string {
	if f.deviceInfo != nil {
		return f.deviceInfo.GetDeviceName()
	}

	return "Not Connected"
}

// Start starts the streaming process (if not already started)
func (f *RadioClient) Start() {
	if !f.streaming {
		log.Debug("Starting streaming")
		f.streaming = true
		f.setStreamState()
	}
}

// Stop stop the streaming process (if started)
func (f *RadioClient) Stop() {
	if f.streaming {
		log.Debug("Stopping")
		f.streaming = false
		f.setStreamState()
	}
}

func (f *RadioClient) setStreamState() {
	if f.streaming {
		if f.iqChannelEnabled {
			go f.iqLoop()
		}
	}
}

func (f *RadioClient) iqLoop() {
	ctx := context.Background()
	iqClient, err := f.client.RXIQ(ctx, &protocol.Session{
    Token: f.deviceInfo.Session,
  })

	if err != nil {
		log.Fatal(err)
	}
	for f.iqChannelEnabled {
		data, err := iqClient.Recv()
		if err != nil {
			log.Error(err)
			f.iqChannelEnabled = false
			break
		}
		cData := data.GetComplexSamples()
		if f.cb != nil {
			f.cb.OnData(cData)
		}
	}
}

// Connect initiates the connection with RadioClient.
// It panics if the connection fails for some reason.
func (f *RadioClient) Connect() {
	if f.routineRunning {
		return
	}

	log.Debug("Trying to connect")

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(f.address, opts...)

	if err != nil {
		log.Fatal(err)
	}

	f.conn = conn

	f.client = protocol.NewRadioServerClient(conn)
	ctx := context.Background()

	log.Debug("Connected, provisioning device.")
	dinf, err := f.client.Provision(ctx, &protocol.DeviceInfo{})
	if err != nil {
		log.Fatal(err)
	}

	log.Debug("Fetching server info")
	sinf, err := f.client.ServerInfo(ctx, &protocol.Empty{})
	if err != nil {
		log.Fatal(err)
	}

  f.deviceInfo = dinf
  f.serverInfo = sinf
}

// Disconnect disconnects from current connected RadioClient.
func (f *RadioClient) Disconnect() {
	log.Debug("Disconnecting")
	f.terminated = true
	f.iqChannelEnabled = false

	if f.conn != nil {
		_ = f.conn.Close()
	}
	f.routineRunning = false
}

// GetSampleRate returns the sample rate of the IQ channel in Hertz
func (f *RadioClient) GetSampleRate() uint32 {
	return f.currentSampleRate
}

// SetSampleRate sets the sample rate of the IQ Channel in Hertz
// Check the available sample rates using GetAvailableSampleRates
// Returns Invalid in case of a invalid value in the input
func (f *RadioClient) SetSampleRate(sampleRate uint32) uint32 {
  f.currentSampleRate = sampleRate
	return f.currentSampleRate
}

// GetCenterFrequency returns the IQ Channel Center Frequency in Hz
func (f *RadioClient) GetCenterFrequency() uint32 {
	return f.iqChannelConfig.CenterFrequency
}

// SetCenterFrequency sets the IQ Channel Center Frequency in Hertz and returns it.
func (f *RadioClient) SetCenterFrequency(centerFrequency uint32) uint32 {
	if f.iqChannelConfig.CenterFrequency != centerFrequency {
		f.iqChannelConfig.CenterFrequency = centerFrequency
	}

	return f.iqChannelConfig.CenterFrequency
}

func (f *RadioClient) SetIQEnabled(iqEnabled bool) {
	f.iqChannelEnabled = iqEnabled
}

// SetCallback sets the callbacks for server data
func (f *RadioClient) SetCallback(cb Callback) {
	f.cb = cb
}

// GetAvailableSampleRates returns a list of available sample rates for the current connection.
func (f *RadioClient) GetAvailableSampleRates() []uint32 {
	return f.availableSampleRates
}

// SetGain sets the gain stage of the server.
// The actual gain in dB varies from device to device.
// Returns Invalid in case of a invalid value in the input
func (f *RadioClient) SetGain(gain uint32) uint32 {
	if f.deviceInfo == nil || gain > f.deviceInfo.MaximumGain {
		return protocol.Invalid
	}
	f.gain = gain

	return gain
}

// GetGain returns the current gain stage of the server.
func (f *RadioClient) GetGain() uint32 {
	return f.gain
}

// endregion
