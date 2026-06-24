# Aqno — desktop companion

Voice-first personal AI companion. This repository is the application skeleton
that turns the Aqno screens into a real desktop app.

```
┌─────────────────────────────────────────────────────────────┐
│  Tauri 2 shell (Rust)                                        │
│  • owns the window, spawns + supervises the sidecar          │
│  • exposes `daemon_url` over IPC, emits `daemon-ready`       │
│                                                              │
│   ┌──────────────────────────┐     ┌──────────────────────┐ │
│   │  SvelteKit webview        │ ⇄  │  aqnod (Go sidecar)   │ │
│   │  • 7 screens + shell      │HTTP │  • REST  /v1/*        │ │
│   │  • design-system tokens   │ SSE │  • SSE   /v1/events   │ │
│   │  • fetch client + stores  │     │  • the always-on work │ │
│   └──────────────────────────┘     └──────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

- **`daemon/`** — the Go daemon (`aqnod`). Runs as a Tauri sidecar on
  `127.0.0.1:8787`, serving the companion's data over REST + a Server-Sent
  Events stream that drives the live presence states. In production this is
  where audio capture, transcription, the local models, the SQLite memory store
  and the VPS links would live; here it returns stable fixtures so the whole UI
  is wired end-to-end.
- **`src/`** — the SvelteKit front (Svelte 5, `adapter-static`, SPA). The seven
  screens — Início, Persona, Agenda, Análise, VPS, Chat, Rede neural — plus the
  persistent shell (sidebar + voice bar). Design-system tokens and components
  are under `src/lib`.
- **`src-tauri/`** — the Tauri 2 Rust shell: window config, sidecar spawn +
  lifecycle, IPC command, capabilities.

## Screens → data

| Screen          | Route      | Daemon endpoint                        |
| --------------- | ---------- | -------------------------------------- |
| Início          | `/`        | `GET /v1/today`                        |
| Criar persona   | `/persona` | _(local state)_                        |
| Agenda          | `/agenda`  | `GET /v1/agenda`                       |
| Briefing diário | `/analise` | `GET /v1/analysis`                     |
| VPS / Infra     | `/vps`     | `GET /v1/vps` · `POST /v1/vps/restart` |
| Chat            | `/chat`    | `GET/POST /v1/chat`                    |
| Rede neural     | `/rede`    | `GET /v1/graph`                        |
| Shell (sidebar) | _all_      | `GET /v1/contexts` · `SSE /v1/events`  |

## Prerequisites

- **Rust** (stable) + the [Tauri 2 system deps](https://tauri.app/start/prerequisites/)
- **Go** ≥ 1.22
- **Node** ≥ 20 + **pnpm** (`corepack enable`)

## Run (desktop)

```bash
pnpm install

# 1. generate real app icons from the bundled orb (once)
pnpm tauri icon src-tauri/app-icon.png

# 2. build the Go sidecar + launch the app
pnpm daemon:sidecar     # compiles daemon → src-tauri/binaries/aqnod-<triple>
pnpm tauri dev
```

`pnpm tauri dev` starts Vite (the `beforeDevCommand`), launches the Rust shell,
which spawns `aqnod`, waits for its `AQNOD_LISTENING <port>` handshake and emits
`daemon-ready` to the webview.

## Run (browser only, no Rust)

The front degrades gracefully to a plain browser — handy for fast UI work:

```bash
# terminal 1 — the daemon
cd daemon && go run .

# terminal 2 — the front (talks to 127.0.0.1:8787 directly)
pnpm dev      # → http://localhost:5173
```

## Production build

```bash
pnpm daemon:sidecar
pnpm tauri build      # bundles the SvelteKit static output + the sidecar
```

## Notes

- **Design system.** Colors, type, spacing, elevation and motion are ported 1:1
  from the Aqno Design System into `src/lib/styles/tokens.css`; the components in
  `src/lib/components` (Presence, VoiceBar, Card, MetricRing, ContextChip,
  Badge, Button, SegmentedControl, ChatBubble, GraphView) mirror that system.
  Fonts: Inter + JetBrains Mono ship via `@fontsource`; add **Geist** for the
  display face to match the brand exactly (the token stack already prefers it).
- **Sidecar naming.** Tauri resolves `binaries/aqnod-<target-triple>`. The
  `pnpm daemon:sidecar` script builds and renames it for your host. See
  `src-tauri/binaries/README.md`.
- **Security.** `tauri.conf.json` scopes the webview CSP `connect-src` to the
  daemon origin, and `capabilities/default.json` only allows spawning the
  `aqnod` sidecar. Tighten the daemon's CORS for production.
- **The reference design** (all 7 screens on one canvas) lives in the parent
  project as `Aqno Telas.dc.html`.
