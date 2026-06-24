//go:build darwin

package voice

import (
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
	"strings"
)

// appleTranscriber drives the aqno-speech helper (Apple SpeechAnalyzer, macOS
// 26+) over stdin/stdout. The helper is gated at runtime via `--probe`.
type appleTranscriber struct{ bin string }

func newAppleTranscriber(bin string) (Transcriber, bool) {
	if bin == "" {
		return nil, false
	}
	out, err := exec.Command(bin, "--probe").Output()
	if err != nil || strings.TrimSpace(string(out)) != "ok" {
		return nil, false
	}
	return &appleTranscriber{bin: bin}, true
}

func (a *appleTranscriber) Name() string    { return "apple-speechanalyzer" }
func (a *appleTranscriber) Available() bool { return a.bin != "" }

func (a *appleTranscriber) TranscribeWAV(ctx context.Context, wav []byte, lang string) (string, error) {
	args := []string{}
	if lang != "" && lang != "auto" {
		args = append(args, "--lang", appleLocale(lang))
	}
	cmd := exec.CommandContext(ctx, a.bin, args...)
	cmd.Stdin = bytes.NewReader(wav)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	var r struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(out.Bytes(), &r); err != nil {
		return strings.TrimSpace(out.String()), nil
	}
	return r.Text, nil
}

// appleLocale maps the app's short language codes to BCP-47 locales.
func appleLocale(lang string) string {
	switch lang {
	case "pt":
		return "pt-BR"
	case "en":
		return "en-US"
	default:
		return lang
	}
}
