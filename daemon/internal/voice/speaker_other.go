//go:build !darwin

package voice

import "context"

// noopSpeaker is the placeholder on non-macOS platforms (Windows/Linux TTS can
// be added behind this interface later).
type noopSpeaker struct{}

// NewSpeaker returns the platform speaker.
func NewSpeaker() Speaker { return noopSpeaker{} }

func (noopSpeaker) Name() string                                 { return "none" }
func (noopSpeaker) Available() bool                              { return false }
func (noopSpeaker) Speak(context.Context, string, string, float64) error { return nil }
