//go:build !whisper_cgo

package voice

// newWhisperTranscriber is a no-op in the default (pure-Go) build. Build the
// daemon with -tags whisper_cgo (and the linked whisper.cpp static lib) to get
// the Metal-accelerated in-process engine.
func newWhisperTranscriber(_ string) (Transcriber, bool) {
	return nil, false
}
