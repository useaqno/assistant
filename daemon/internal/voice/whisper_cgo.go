//go:build whisper_cgo

// whisper.cpp in-process engine (Metal-accelerated on Apple Silicon).
//
// Build prerequisites (see scripts/build-sidecar.mjs and docs WS4 design):
//   1. Vendor whisper.cpp at third_party/whisper.cpp (pinned tag).
//   2. Build the static lib with Metal embedded:
//        cmake -S third_party/whisper.cpp -B build_go -DBUILD_SHARED_LIBS=OFF \
//          -DGGML_METAL=ON -DGGML_METAL_EMBED_LIBRARY=ON
//        cmake --build build_go --target whisper -j
//   3. go build -tags whisper_cgo with C_INCLUDE_PATH/LIBRARY_PATH pointing at
//      the headers and the build_go/*/*.a archives.
package voice

import (
	"context"
	"encoding/binary"
	"errors"
	"os"
	"strings"

	whisper "github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
)

// decodeWAVToFloat32 reads a 16-bit PCM WAV into normalized float32 samples.
// whisper.cpp expects 16 kHz mono; the daemon writes WAVs in that format.
func decodeWAVToFloat32(wav []byte) ([]float32, error) {
	if len(wav) < 44 || string(wav[0:4]) != "RIFF" {
		return nil, errors.New("invalid WAV")
	}
	// Locate the data chunk.
	i := 12
	var data []byte
	for i+8 <= len(wav) {
		id := string(wav[i : i+4])
		sz := int(binary.LittleEndian.Uint32(wav[i+4 : i+8]))
		body := i + 8
		if id == "data" {
			end := body + sz
			if end > len(wav) {
				end = len(wav)
			}
			data = wav[body:end]
			break
		}
		i = body + sz
	}
	if data == nil {
		return nil, errors.New("no data chunk")
	}
	n := len(data) / 2
	out := make([]float32, n)
	for j := 0; j < n; j++ {
		s := int16(binary.LittleEndian.Uint16(data[j*2 : j*2+2]))
		out[j] = float32(s) / 32768
	}
	return out, nil
}

type whisperTranscriber struct {
	modelPath string
	model     whisper.Model
}

// newWhisperTranscriber loads the model if the path exists.
func newWhisperTranscriber(modelPath string) (Transcriber, bool) {
	if modelPath == "" {
		return nil, false
	}
	if _, err := os.Stat(modelPath); err != nil {
		return nil, false
	}
	model, err := whisper.New(modelPath)
	if err != nil {
		return nil, false
	}
	return &whisperTranscriber{modelPath: modelPath, model: model}, true
}

func (w *whisperTranscriber) Name() string     { return "whisper.cpp" }
func (w *whisperTranscriber) Available() bool   { return w.model != nil }

func (w *whisperTranscriber) TranscribeWAV(_ context.Context, wav []byte, lang string) (string, error) {
	samples, err := decodeWAVToFloat32(wav)
	if err != nil {
		return "", err
	}
	ctx, err := w.model.NewContext()
	if err != nil {
		return "", err
	}
	if lang != "" && lang != "auto" {
		_ = ctx.SetLanguage(lang)
	}
	if err := ctx.Process(samples, nil, nil, nil); err != nil {
		return "", err
	}
	var b strings.Builder
	for {
		seg, err := ctx.NextSegment()
		if err != nil {
			break
		}
		b.WriteString(seg.Text)
	}
	return strings.TrimSpace(b.String()), nil
}
