package voice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

// serverTranscriber talks to a whisper.cpp HTTP server (the `whisper-server`
// example) via its OpenAI-compatible /inference multipart endpoint. Pure Go —
// no cgo — so the daemon keeps building as a single binary.
type serverTranscriber struct {
	base string
}

func (s *serverTranscriber) Name() string    { return "whisper-server" }
func (s *serverTranscriber) Available() bool { return s.base != "" }

func (s *serverTranscriber) TranscribeWAV(ctx context.Context, wav []byte, lang string) (string, error) {
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, err := mw.CreateFormFile("file", "audio.wav")
	if err != nil {
		return "", err
	}
	if _, err := fw.Write(wav); err != nil {
		return "", err
	}
	_ = mw.WriteField("response_format", "json")
	if lang != "" && lang != "auto" {
		_ = mw.WriteField("language", lang)
	}
	mw.Close()

	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		strings.TrimRight(s.base, "/")+"/inference", &body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("whisper-server: %d %s", res.StatusCode, string(data))
	}
	var out struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return strings.TrimSpace(string(data)), nil
	}
	return strings.TrimSpace(out.Text), nil
}
