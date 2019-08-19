package DSP

import (
	"github.com/quan-to/slog"
	"github.com/racerxdl/go.fifo"
	"runtime"
	"sync"
	"time"
)

var cgLog = slog.Scope("ChannelGenerator")

const maxFifoSize = 4096

type OnIQSamples func(samples []complex64)

type ChannelGenerator struct {
	sync.Mutex

	inputFifo     *fifo.Queue
	running       bool
	settingsMutex sync.Mutex

	iqEnabled      bool

	onIQSamples    OnIQSamples
	updateChannel  chan bool

	syncSampleInput *sync.Cond
}

func CreateChannelGenerator() *ChannelGenerator {
	var cg = &ChannelGenerator{
		Mutex:         sync.Mutex{},
		inputFifo:     fifo.NewQueue(),
		settingsMutex: sync.Mutex{},
		updateChannel: make(chan bool),
	}

	cg.syncSampleInput = sync.NewCond(cg)
	return cg
}

func (cg *ChannelGenerator) routine() {
	for cg.running {
		go func() {
			<-time.After(1 * time.Second)
			cg.syncSampleInput.Broadcast()
		}()
		cg.syncSampleInput.L.Lock()
		cg.syncSampleInput.Wait()
		cg.doWork()
		cg.syncSampleInput.L.Unlock()

		if !cg.running {
			break
		}
		runtime.Gosched()
	}
	cgLog.Debug("Cleaning fifo")
	for i := 0; i < cg.inputFifo.Len(); i++ {
		cg.inputFifo.Next()
	}
	cgLog.Debug("Done")
}

func (cg *ChannelGenerator) doWork() {
	cg.settingsMutex.Lock()
	for cg.inputFifo.Len() > 0 {
		var samples = cg.inputFifo.Next().([]complex64)
		if cg.iqEnabled {
			cg.processIQ(samples)
		}
	}
	cg.settingsMutex.Unlock()
}

func (cg *ChannelGenerator) processIQ(samples []complex64) {
	if cg.onIQSamples != nil {
		cg.onIQSamples(samples)
	}
}

func (cg *ChannelGenerator) notify() {
	cg.syncSampleInput.Broadcast()
}

func (cg *ChannelGenerator) Start() {
	if !cg.running {
		cgLog.Info("Starting Channel Generator")
		cg.running = true
		go cg.routine()
		//go func() {
		//	for cg.running {
		//		<-time.After(1 * time.Second)
		//		cgLog.Debug("Fifo Usage: %d", cg.inputFifo.UnsafeLen())
		//	}
		//}()
	}
}

func (cg *ChannelGenerator) Stop() {
	if cg.running {
		cgLog.Info("Stopping")
		cg.running = false
		cg.notify()
	}
}

func (cg *ChannelGenerator) StartIQ() {
	cg.settingsMutex.Lock()
	cgLog.Info("Enabling IQ")
	cg.iqEnabled = true
	cg.settingsMutex.Unlock()
}

func (cg *ChannelGenerator) StopIQ() {
	cg.settingsMutex.Lock()
	cgLog.Info("Disabling IQ")
	cg.iqEnabled = false
	cg.settingsMutex.Unlock()
}

func (cg *ChannelGenerator) PushSamples(samples []complex64) {
	if !cg.running {
		return
	}

	var fifoLength = cg.inputFifo.Len()

	if maxFifoSize <= fifoLength {
		cgLog.Debug("Fifo Overflowing!")
		return
	}

	cg.inputFifo.Add(samples)

	cg.notify()
}

func (cg *ChannelGenerator) SetOnIQ(cb OnIQSamples) {
	cg.onIQSamples = cb
}

func (cg *ChannelGenerator) IQRunning() bool {
	return cg.iqEnabled
}
