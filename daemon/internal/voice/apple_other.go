//go:build !darwin

package voice

// newAppleTranscriber is unavailable off macOS (Apple SpeechAnalyzer is
// macOS/iOS-only).
func newAppleTranscriber(_ string) (Transcriber, bool) {
	return nil, false
}
