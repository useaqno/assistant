//go:build darwin

package voice

import (
	"context"
	"os/exec"
	"strconv"
)

// saySpeaker uses the built-in macOS `say` command (zero dependencies). A neural
// TTS (e.g. Piper) can be added later behind the same interface.
type saySpeaker struct{}

// NewSpeaker returns the platform speaker.
func NewSpeaker() Speaker { return saySpeaker{} }

func (saySpeaker) Name() string    { return "macos-say" }
func (saySpeaker) Available() bool {
	_, err := exec.LookPath("say")
	return err == nil
}

func (saySpeaker) Speak(ctx context.Context, text, voice string, rate float64) error {
	args := []string{}
	if voice != "" {
		args = append(args, "-v", voice)
	}
	if rate > 0 {
		// `say` rate is words/min; map 1.0x ≈ 180 wpm.
		args = append(args, "-r", strconv.Itoa(int(180*rate)))
	}
	args = append(args, text)
	return exec.CommandContext(ctx, "say", args...).Run()
}
