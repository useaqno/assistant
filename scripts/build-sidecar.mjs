// Builds the Go daemon and names it the way Tauri expects a sidecar:
//   src-tauri/binaries/aqnod-<rust-target-triple>[.exe]
//
// Run via `pnpm daemon:sidecar`. The `pnpm app:dev` script runs this first so
// `tauri dev` always finds an up-to-date binary.
//
// Target selection:
//   - default: the host Rust triple (what `tauri dev`/`tauri build` use).
//   - AQNO_TARGET=<rust-triple>: cross-build the pure-Go daemon for that triple
//     (used by the Makefile's multi-platform builds; CGO disabled).
//
// Default build is pure Go (no cgo): voice STT uses the whisper.cpp HTTP server
// (AQNO_WHISPER_SERVER) when configured; TTS uses the OS speech engine.
//
// Optional in-process whisper.cpp (Metal-accelerated) engine — host only:
//   AQNO_WHISPER_CGO=1 pnpm daemon:sidecar
// Requires cmake and a vendored whisper.cpp at third_party/whisper.cpp.
import { execSync } from 'node:child_process'
import { cpSync, mkdirSync, existsSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { dirname, resolve } from 'node:path'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const useCgo = process.env.AQNO_WHISPER_CGO === '1'

// rust-triple -> Go {GOOS, GOARCH}
const TRIPLE_TO_GO = {
  'aarch64-apple-darwin': { GOOS: 'darwin', GOARCH: 'arm64' },
  'x86_64-apple-darwin': { GOOS: 'darwin', GOARCH: 'amd64' },
  'x86_64-pc-windows-msvc': { GOOS: 'windows', GOARCH: 'amd64' },
  'aarch64-pc-windows-msvc': { GOOS: 'windows', GOARCH: 'arm64' },
  'x86_64-unknown-linux-gnu': { GOOS: 'linux', GOARCH: 'amd64' },
  'aarch64-unknown-linux-gnu': { GOOS: 'linux', GOARCH: 'arm64' }
}

function hostTriple() {
  const out = execSync('rustc -Vv', { encoding: 'utf8' })
  const line = out.split('\n').find((l) => l.startsWith('host:'))
  if (!line) throw new Error('Could not determine rust host triple — is rustc installed?')
  return line.split(':')[1].trim()
}

const triple = process.env.AQNO_TARGET || hostTriple()
const goTarget = TRIPLE_TO_GO[triple]
const isCross = !!process.env.AQNO_TARGET && triple !== hostTriple()
const ext = triple.includes('windows') ? '.exe' : ''

const binDir = resolve(root, 'src-tauri/binaries')
const daemonDir = resolve(root, 'daemon')
mkdirSync(binDir, { recursive: true })

const tripled = resolve(binDir, `aqnod-${triple}${ext}`)
let env = { ...process.env }
let buildArgs = '.'

if (useCgo) {
  if (isCross) {
    throw new Error('AQNO_WHISPER_CGO cannot be combined with cross-compilation (AQNO_TARGET).')
  }
  const wcpp = resolve(root, 'third_party/whisper.cpp')
  if (!existsSync(wcpp)) {
    throw new Error(
      `AQNO_WHISPER_CGO=1 but ${wcpp} is missing. Vendor whisper.cpp there (pinned tag) first.`
    )
  }
  const buildGo = resolve(wcpp, 'build_go')
  const metal = process.platform === 'darwin'
  console.log('› building whisper.cpp static lib (cmake)…')
  execSync(
    `cmake -S "${wcpp}" -B "${buildGo}" -DCMAKE_BUILD_TYPE=Release -DBUILD_SHARED_LIBS=OFF ` +
      `-DWHISPER_BUILD_TESTS=OFF -DWHISPER_BUILD_EXAMPLES=OFF ` +
      (metal ? '-DGGML_METAL=ON -DGGML_METAL_EMBED_LIBRARY=ON' : ''),
    { stdio: 'inherit' }
  )
  execSync(`cmake --build "${buildGo}" --target whisper -j`, { stdio: 'inherit' })
  env = {
    ...env,
    CGO_ENABLED: '1',
    C_INCLUDE_PATH: `${resolve(wcpp, 'include')}:${resolve(wcpp, 'ggml/include')}`,
    LIBRARY_PATH: [
      resolve(buildGo, 'src'),
      resolve(buildGo, 'ggml/src'),
      resolve(buildGo, 'ggml/src/ggml-metal'),
      resolve(buildGo, 'ggml/src/ggml-blas')
    ].join(':')
  }
  buildArgs = '-tags whisper_cgo .'
  console.log('› building aqnod with the in-process whisper.cpp engine')
} else {
  // Pure-Go build (the modernc.org/sqlite driver needs no cgo).
  env = { ...env, CGO_ENABLED: '0' }
  if (goTarget) {
    env.GOOS = goTarget.GOOS
    env.GOARCH = goTarget.GOARCH
  }
  console.log(`› building aqnod (pure Go) for ${triple}`)
}

execSync(`go build -o "${tripled}" ${buildArgs}`, { cwd: daemonDir, stdio: 'inherit', env })
console.log(`✓ sidecar ready → src-tauri/binaries/aqnod-${triple}${ext}`)

// macOS only: build the optional Apple SpeechAnalyzer helper (aqno-speech).
buildSpeechHelper()

function buildSpeechHelper() {
  if (process.platform !== 'darwin' || !triple.includes('apple-darwin')) return
  const helperDir = resolve(root, 'helpers/aqno-speech')
  if (!existsSync(helperDir)) return
  try {
    execSync('swift --version', { stdio: 'ignore' })
  } catch {
    console.log('› skipping aqno-speech (swift not found)')
    return
  }
  const arch = triple.startsWith('aarch64') ? 'arm64' : 'x86_64'
  console.log('› building aqno-speech (Apple SpeechAnalyzer helper)…')
  try {
    execSync(`swift build -c release --arch ${arch}`, { cwd: helperDir, stdio: 'inherit' })
    const built = resolve(helperDir, `.build/${arch}-apple-macosx/release/aqno-speech`)
    const fallback = resolve(helperDir, '.build/release/aqno-speech')
    cpSync(existsSync(built) ? built : fallback, resolve(binDir, `aqno-speech-${triple}`))
    console.log(`✓ helper ready → src-tauri/binaries/aqno-speech-${triple}`)
  } catch (e) {
    console.log(`› aqno-speech build failed (optional engine): ${e.message}`)
  }
}
