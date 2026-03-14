package main

import (
	"bytes"
	"encoding/binary"
	"sync"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/speaker"
)

type Player struct {
	buffer     []byte
	bufferCond *sync.Cond
	mu         sync.Mutex
}

func NewPlayer() (*Player, error) {
	err := speaker.Init(
		beep.SampleRate(24000),                   // Match the 24kHz sample rate of pcm16 data
		beep.SampleRate(24000).N(time.Second/10), // Buffer size
	)
	if err != nil {
		return nil, err
	}

	return &Player{
		buffer:     make([]byte, 0),
		bufferCond: sync.NewCond(&sync.Mutex{}),
	}, nil
}

func (p *Player) AddChunk(chunk []byte) {
	p.bufferCond.L.Lock()
	defer p.bufferCond.L.Unlock()
	p.buffer = append(p.buffer, chunk...)
	p.bufferCond.Signal() // Notify that new data is available
}

func (p *Player) Play() {
	streamer := beep.StreamerFunc(func(samples [][2]float64) (n int, ok bool) {
		p.bufferCond.L.Lock()
		defer p.bufferCond.L.Unlock()

		for i := range samples {
			// Wait until there’s enough data in the buffer for a sample
			for len(p.buffer) < 2 {
				p.bufferCond.Wait()
			}

			var sample int16
			err := binary.Read(bytes.NewReader(p.buffer[:2]), binary.LittleEndian, &sample)
			if err != nil {
				return i, false
			}

			// Convert PCM sample to normalized float64 range for beep
			floatSample := float64(sample) / 32768.0
			samples[i][0] = floatSample // Left channel (mono)
			samples[i][1] = floatSample // Right channel (duplicate mono)

			// Remove processed bytes from the buffer
			p.buffer = p.buffer[2:]
		}

		return len(samples), true
	})

	speaker.Play(streamer)
}

func (p *Player) Close() {
	// Stop the speaker and release audio resources
	speaker.Close()

	// Optional: Clear buffer, if you want a clean slate on next start
	p.mu.Lock()
	defer p.mu.Unlock()
	p.buffer = nil
}
