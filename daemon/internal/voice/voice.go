// Package voice is the speech pipeline: a Transcriber (STT) and Speaker (TTS)
// behind interfaces so whisper.cpp (cgo), a whisper-server sidecar, or a future
// engine are interchangeable. See docs/context.md §7 and the WS4 design notes.
//
// Engine selection at runtime:
//   - whisper.cpp via cgo (build with -tags whisper_cgo, Metal-accelerated) when
//     a model is present;
//   - otherwise a whisper.cpp HTTP server (AQNO_WHISPER_SERVER) if configured;
//   - otherwise unavailable (the UI guides the user to download a model / use
//     the browser STT in dev).
package voice

import (
	"bytes"
	"context"
	"encoding/binary"
	"math"
)

// Transcriber turns audio into text. Lang is a BCP-ish code ("pt", "en", "auto").
type Transcriber interface {
	Name() string
	Available() bool
	// TranscribeWAV transcribes a 16-bit PCM WAV (any sample rate; engines
	// resample as needed) and returns the recognized text.
	TranscribeWAV(ctx context.Context, wav []byte, lang string) (string, error)
}

// Speaker turns text into spoken audio on the local device.
type Speaker interface {
	Name() string
	Available() bool
	Speak(ctx context.Context, text, voice string, rate float64) error
}

// Options configures transcriber construction.
type Options struct {
	ModelPath string // local whisper model (for the cgo engine)
	ServerURL string // whisper.cpp server base URL (for the server engine)
}

// NewTranscriber picks the best available engine (cgo > server > unavailable).
func NewTranscriber(opts Options) Transcriber {
	if t, ok := newWhisperTranscriber(opts.ModelPath); ok {
		return t
	}
	if opts.ServerURL != "" {
		return &serverTranscriber{base: opts.ServerURL}
	}
	return unavailable{}
}

// unavailable is the no-engine fallback.
type unavailable struct{}

func (unavailable) Name() string    { return "none" }
func (unavailable) Available() bool { return false }
func (unavailable) TranscribeWAV(context.Context, []byte, string) (string, error) {
	return "", ErrNoEngine
}

// ErrNoEngine signals that no STT backend is configured.
var ErrNoEngine = errStr("nenhum motor de transcrição disponível — baixe um modelo em Ajustes ou configure AQNO_WHISPER_SERVER")

type errStr string

func (e errStr) Error() string { return string(e) }

// FloatsToWAV encodes float32 [-1,1] mono PCM as a 16-bit WAV at sampleRate.
func FloatsToWAV(samples []float32, sampleRate int) []byte {
	var buf bytes.Buffer
	dataLen := len(samples) * 2
	// RIFF header
	buf.WriteString("RIFF")
	binary.Write(&buf, binary.LittleEndian, uint32(36+dataLen))
	buf.WriteString("WAVE")
	// fmt chunk
	buf.WriteString("fmt ")
	binary.Write(&buf, binary.LittleEndian, uint32(16))
	binary.Write(&buf, binary.LittleEndian, uint16(1)) // PCM
	binary.Write(&buf, binary.LittleEndian, uint16(1)) // mono
	binary.Write(&buf, binary.LittleEndian, uint32(sampleRate))
	binary.Write(&buf, binary.LittleEndian, uint32(sampleRate*2)) // byte rate
	binary.Write(&buf, binary.LittleEndian, uint16(2))            // block align
	binary.Write(&buf, binary.LittleEndian, uint16(16))           // bits per sample
	// data chunk
	buf.WriteString("data")
	binary.Write(&buf, binary.LittleEndian, uint32(dataLen))
	for _, s := range samples {
		v := s
		if v > 1 {
			v = 1
		} else if v < -1 {
			v = -1
		}
		binary.Write(&buf, binary.LittleEndian, int16(math.Round(float64(v)*32767)))
	}
	return buf.Bytes()
}
