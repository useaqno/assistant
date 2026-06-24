package voice

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"aqnod/internal/modelmanager"
	"aqnod/internal/store"
)

// Service ties the model manager, transcriber and speaker together, reading
// runtime preferences (tier, language, TTS voice) from the store.
type Service struct {
	st      *store.Store
	mm      *modelmanager.Manager
	speaker Speaker

	mu          sync.RWMutex
	transcriber Transcriber
}

// NewService wires the voice pipeline.
func NewService(st *store.Store) *Service {
	dir := modelsDir()
	s := &Service{
		st:      st,
		mm:      modelmanager.New(dir, st.ConfigVal("voice.model_mirror", "")),
		speaker: NewSpeaker(),
	}
	s.rebuild()
	return s
}

func modelsDir() string {
	base, err := store.DataDir()
	if err != nil {
		base = os.TempDir()
	}
	return filepath.Join(base, "models")
}

// rebuild reselects the transcriber engine (after a model download or config change).
func (s *Service) rebuild() {
	tier := modelmanager.Tier(s.st.ConfigVal("voice.model_tier", "small"))
	modelPath := ""
	if p, err := s.mm.Path(tier); err == nil {
		if st := s.mm.Status(tier); st.Present {
			modelPath = p
		}
	}
	t := NewTranscriber(Options{
		ModelPath: modelPath,
		ServerURL: os.Getenv("AQNO_WHISPER_SERVER"),
	})
	s.mu.Lock()
	s.transcriber = t
	s.mu.Unlock()
}

func (s *Service) tx() Transcriber {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.transcriber
}

// Models returns download/verification status per tier.
func (s *Service) Models() []modelmanager.Status { return s.mm.AllStatus() }

// Download fetches a model tier, then reselects the engine. Progress is streamed.
func (s *Service) Download(ctx context.Context, tier string, prog chan<- modelmanager.Progress) error {
	_, err := s.mm.Ensure(ctx, modelmanager.Tier(tier), prog)
	if err == nil {
		s.rebuild()
	}
	return err
}

// Engine reports the active transcriber name and availability.
func (s *Service) Engine() (string, bool) {
	t := s.tx()
	return t.Name(), t.Available()
}

// Transcribe runs STT on a 16-bit PCM WAV.
func (s *Service) Transcribe(ctx context.Context, wav []byte) (string, error) {
	lang := s.st.ConfigVal("voice.stt_lang", "pt")
	return s.tx().TranscribeWAV(ctx, wav, lang)
}

// TranscribeFloats runs STT on float32 mono PCM at sampleRate.
func (s *Service) TranscribeFloats(ctx context.Context, samples []float32, sampleRate int) (string, error) {
	return s.Transcribe(ctx, FloatsToWAV(samples, sampleRate))
}

// Speak speaks text using the configured TTS voice and speed.
func (s *Service) Speak(ctx context.Context, text string) error {
	if !s.speaker.Available() {
		return nil
	}
	voice := s.st.ConfigVal("voice.tts_voice", "")
	speed := 1.0
	if v := s.st.ConfigVal("voice.speed", "1.0"); v != "" {
		_, _ = fmt.Sscanf(v, "%g", &speed)
	}
	return s.speaker.Speak(ctx, text, voice, speed)
}

// SpeakerAvailable reports TTS availability.
func (s *Service) SpeakerAvailable() bool { return s.speaker.Available() }
