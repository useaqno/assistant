# Sidecar binaries

Tauri bundles the Go daemon (`aqnod`) as an external binary. It must be named
with the Rust **host target triple** so Tauri can pick the right one per platform:

```
aqnod-aarch64-apple-darwin
aqnod-x86_64-apple-darwin
aqnod-x86_64-unknown-linux-gnu
aqnod-x86_64-pc-windows-msvc.exe
```

Generate it with:

```
pnpm daemon:sidecar
```

(That script runs `go build` and copies the result to the triple-suffixed name.)
These files are git-ignored — they are build artifacts.
