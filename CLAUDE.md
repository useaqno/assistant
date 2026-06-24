# Aqno — Claude Code Guide

## What this is

Desktop AI companion (voice-first) built with Tauri 2 + SvelteKit + Go daemon.
Three-layer architecture: Rust shell → Svelte 5 webview ↔ Go HTTP/SSE sidecar at `127.0.0.1:8787`.

## Dev commands

```bash
# Frontend only (fastest iteration)
pnpm dev

# Full desktop app
pnpm daemon:sidecar && pnpm tauri dev

# Go daemon standalone
cd daemon && go run .

# Type-check
pnpm check

# Lint
pnpm lint
pnpm lint:fix

# Format
pnpm format
```

## Architecture

```
src-tauri/          Rust shell — spawns aqnod, manages window
src/routes/         SvelteKit screens (7 routes)
src/lib/
  api.ts            Typed REST client
  tauri.ts          IPC bridge (getDaemonUrl, onDaemonReady, SSE)
  stores/           Svelte stores (presence, voice)
  components/       Shared UI components
  styles/tokens.css Design tokens — ONLY source of color/spacing/type values
daemon/
  main.go           HTTP handlers + SSE stream
  data.go           Mock fixtures (replace with SQLite + real logic)
```

## Code Standards

### Conventional Commits (enforced by commitlint)

Format: `<type>(<scope>): <subject>`

Types:

- `feat` — new feature
- `fix` — bug fix
- `refactor` — restructure without behavior change
- `perf` — performance improvement
- `style` — formatting only (no logic change)
- `test` — add/fix tests
- `docs` — documentation
- `chore` — build, deps, tooling
- `ci` — CI/CD changes
- `revert` — revert a commit

Scopes (optional): `ui`, `daemon`, `tauri`, `api`, `voice`, `agenda`, `chat`, `vps`, `graph`, `deps`

Good examples:

```
feat(chat): add message persistence to SQLite
fix(daemon): handle SSE client disconnect without panic
refactor(ui): extract MetricRing into standalone component
chore(deps): upgrade Svelte to 5.2
```

Rules:

- Subject ≤ 72 chars, lowercase, no period at end
- Body wrapped at 100 chars
- Use imperative mood ("add", not "added")

### TypeScript / Svelte 5

- **No `any`** — use `unknown` + type guards or proper types
- **Svelte 5 runes** — `$state`, `$derived`, `$effect` for component state; never `writable`/`readable` inside components
- **Type imports** — always `import type { Foo }` for type-only imports
- **No magic values** — colors/spacing from `tokens.css` only
- **Short functions** — if a function needs a comment to explain what it does, split or rename it
- **No barrel re-exports** unless the module is a genuine public API

### Go (daemon)

- `gofmt` always (enforced by CI)
- Errors are values — never ignore `err`
- Keep handlers thin: parse → call service → respond
- No global mutable state outside the server struct

### General

- **No TODO comments** in committed code — open a GitHub issue instead
- **No commented-out code** — delete it; git history keeps it
- **One concern per file** — a Svelte component does one thing

## Design tokens

All visual values live in `src/lib/styles/tokens.css`.
Never hardcode hex colors, px sizes, or font weights inline.

```css
/* correct */
background: var(--color-surface-1);
gap: var(--space-4);

/* wrong */
background: #1a1a2e;
gap: 16px;
```

## File naming

| Type              | Convention           | Example              |
| ----------------- | -------------------- | -------------------- |
| Svelte components | PascalCase           | `VoiceBar.svelte`    |
| TS modules        | camelCase            | `api.ts`, `tauri.ts` |
| Svelte routes     | SvelteKit convention | `+page.svelte`       |
| Go files          | snake_case           | `main.go`, `data.go` |
