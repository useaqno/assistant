package httpapi

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/coder/websocket"

	"aqnod/internal/modelmanager"
)

func (s *Server) handleVoiceEngine(w http.ResponseWriter, _ *http.Request) {
	if s.deps.Voice == nil {
		writeJSON(w, http.StatusOK, map[string]any{"active": "none", "available": false, "apple": false})
		return
	}
	name, ok := s.deps.Voice.Engine()
	writeJSON(w, http.StatusOK, map[string]any{
		"active":    name,
		"available": ok,
		"apple":     s.deps.Voice.AppleAvailable(),
	})
}

func (s *Server) handleVoiceModels(w http.ResponseWriter, _ *http.Request) {
	if s.deps.Voice == nil {
		writeJSON(w, http.StatusOK, []any{})
		return
	}
	writeJSON(w, http.StatusOK, s.deps.Voice.Models())
}

func (s *Server) handleDownloadModel(w http.ResponseWriter, r *http.Request) {
	if s.deps.Voice == nil {
		writeErr(w, http.StatusServiceUnavailable, "voz indisponível")
		return
	}
	tier := r.PathValue("tier")
	// Models are large; download in the background and stream progress over SSE.
	go func() {
		prog := make(chan modelmanager.Progress, 8)
		go func() {
			for p := range prog {
				s.deps.Hub.Broadcast("voice.model", map[string]any{
					"tier": string(p.Tier), "downloaded": p.Downloaded, "total": p.Total,
					"done": p.Done, "error": errString(p.Err),
				})
			}
		}()
		_ = s.deps.Voice.Download(context.Background(), tier, prog)
	}()
	writeJSON(w, http.StatusAccepted, map[string]any{"ok": true, "status": "downloading", "tier": tier})
}

func (s *Server) handleSpeak(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Text string `json:"text"`
	}
	if err := readJSON(r, &b); err != nil || b.Text == "" {
		writeErr(w, http.StatusBadRequest, "text required")
		return
	}
	if s.deps.Voice == nil || !s.deps.Voice.SpeakerAvailable() {
		writeJSON(w, http.StatusOK, map[string]any{"ok": false, "reason": "tts indisponível"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	if err := s.deps.Voice.Speak(ctx, b.Text); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleTranscribe(w http.ResponseWriter, r *http.Request) {
	if s.deps.Voice == nil {
		writeErr(w, http.StatusServiceUnavailable, "voz indisponível")
		return
	}
	wav, err := io.ReadAll(io.LimitReader(r.Body, 50<<20)) // 50 MB cap
	if err != nil {
		writeErr(w, http.StatusBadRequest, "audio inválido")
		return
	}
	text, err := s.deps.Voice.Transcribe(r.Context(), wav)
	if err != nil {
		writeErr(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"text": text})
}

// handleVoiceStream accepts binary float32 PCM frames over WebSocket and a JSON
// control channel; on "stop" it transcribes, runs the intent, and replies.
func (s *Server) handleVoiceStream(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		return
	}
	defer c.Close(websocket.StatusNormalClosure, "")
	ctx := r.Context()

	var samples []float32
	sampleRate := 16000
	lang := ""

	writeEvent := func(ev string, payload any) {
		b, _ := json.Marshal(map[string]any{"type": ev, "data": payload})
		_ = c.Write(ctx, websocket.MessageText, b)
	}

	for {
		typ, data, err := c.Read(ctx)
		if err != nil {
			return
		}
		switch typ {
		case websocket.MessageBinary:
			samples = append(samples, decodeFloat32LE(data)...)
		case websocket.MessageText:
			var ctrl struct {
				Type       string `json:"type"`
				SampleRate int    `json:"sampleRate"`
				Lang       string `json:"lang"`
				Context    string `json:"context"`
			}
			_ = json.Unmarshal(data, &ctrl)
			switch ctrl.Type {
			case "start":
				samples = samples[:0]
				if ctrl.SampleRate > 0 {
					sampleRate = ctrl.SampleRate
				}
				lang = ctrl.Lang
			case "stop":
				_ = lang
				if s.deps.Voice == nil || len(samples) == 0 {
					writeEvent("final", map[string]string{"text": ""})
					continue
				}
				text, terr := s.deps.Voice.TranscribeFloats(ctx, samples, sampleRate)
				if terr != nil {
					writeEvent("error", map[string]string{"message": terr.Error()})
					continue
				}
				writeEvent("final", map[string]string{"text": text})
				if text != "" {
					reply, _ := s.brain().Intent(text, ctrl.Context)
					writeEvent("reply", reply)
					go s.deps.Voice.Speak(context.Background(), reply.Text)
				}
				samples = samples[:0]
			}
		}
	}
}

func decodeFloat32LE(b []byte) []float32 {
	n := len(b) / 4
	out := make([]float32, n)
	for i := 0; i < n; i++ {
		out[i] = math.Float32frombits(binary.LittleEndian.Uint32(b[i*4 : i*4+4]))
	}
	return out
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
