package client

import (
  "context"
	"encoding/json"
	"github.com/quan-to/slog"
	"github.com/luigifreitas/radioserver/protocol"
  "google.golang.org/grpc/encoding/gzip"
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
	deviceState    *protocol.DeviceState
  session        *protocol.Session
  ctx            context.Context
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
    currentSampleRate:     600000,
    ctx:                   context.Background(),
  }
}

// region Public Methods

// GetName returns the name of the active device in RadioClient
func (f *RadioClient) GetName() string {
	return f.deviceState.Info.Name.String()
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
	iqClient, err := f.client.RXIQ(f.ctx, f.session)

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
  opts = append(opts, grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name)))
	conn, err := grpc.Dial(f.address, opts...)

	if err != nil {
		log.Fatal(err)
	}

	f.conn = conn

	f.client = protocol.NewRadioServerClient(conn)

  log.Debug("Connected, listing devices.")
	dls, err := f.client.List(f.ctx, &protocol.Empty{})
	if err != nil {
		log.Fatal(err)
	}

  i := protocol.DeviceState{
    Info: dls.Devices[0],
    Config: &protocol.DeviceConfig{
      SampleRate: float32(f.currentSampleRate),
      Oversample: 16,
    },
  }

  i.Config.RXC = append(i.Config.RXC,
    &protocol.ChannelConfig{
      CenterFrequency: 96.9e6,
      NormalizedGain: 0.5,
      Antenna: "LNAW",
    })

  deviceProv, _ := json.MarshalIndent(i, "", "   ")
	log.Info("Provisioning Device: %s", deviceProv)

  session, err := f.client.Provision(f.ctx, &i)
	if err != nil {
		log.Fatal(err)
	}

  f.session = session
  f.deviceState = &i

	log.Debug("Fetching server info")
	sinf, err := f.client.ServerInfo(f.ctx, &protocol.Empty{})
	if err != nil {
		log.Fatal(err)
	}

  f.serverInfo = sinf
}

func (f *RadioClient) ChangeFrequency(cf float32) {
  f.deviceState.Config.RXC[0].CenterFrequency = cf

  deviceProv, _ := json.MarshalIndent(f.deviceState, "", "   ")
	log.Info("Retuning Device: %s", deviceProv)

  _, err := f.client.Tune(f.ctx, &protocol.DeviceTune{
    Session: f.session,
    Config: f.deviceState.Config,
  })
	if err != nil {
		log.Fatal(err)
	}
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
	return uint32(f.iqChannelConfig.CenterFrequency)
}

// SetCenterFrequency sets the IQ Channel Center Frequency in Hertz and returns it.
func (f *RadioClient) SetCenterFrequency(centerFrequency uint32) uint32 {
	return 0//f.iqChannelConfig.CenterFrequency
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
	f.gain = gain
	return gain
}

// GetGain returns the current gain stage of the server.
func (f *RadioClient) GetGain() uint32 {
	return f.gain
}

// endregion
