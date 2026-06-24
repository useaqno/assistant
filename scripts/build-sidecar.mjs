// Builds the Go daemon and names it the way Tauri expects a sidecar:
//   src-tauri/binaries/aqnod-<rust-host-target-triple>[.exe]
//
// Run via `pnpm daemon:sidecar`. The `pnpm app:dev` script runs this first so
// `tauri dev` always finds an up-to-date binary.
//
// Default build is pure Go (no cgo): voice STT uses the whisper.cpp HTTP server
// (AQNO_WHISPER_SERVER) when configured; TTS uses the OS speech engine.
//
// Optional in-process whisper.cpp (Metal-accelerated) engine:
//   AQNO_WHISPER_CGO=1 pnpm daemon:sidecar
// Requires cmake and a vendored whisper.cpp at third_party/whisper.cpp. The
// script builds the static lib (Metal embedded) and links it via cgo.
import { execSync } from 'node:child_process'
import { cpSync, mkdirSync, existsSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { dirname, resolve } from 'node:path'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const ext = process.platform === 'win32' ? '.exe' : ''
const useCgo = process.env.AQNO_WHISPER_CGO === '1'

function hostTriple() {
  const out = execSync('rustc -Vv', { encoding: 'utf8' })
  const line = out.split('\n').find((l) => l.startsWith('host:'))
  if (!line) throw new Error('Could not determine rust host triple — is rustc installed?')
  return line.split(':')[1].trim()
}

const triple = hostTriple()
const binDir = resolve(root, 'src-tauri/binaries')
const daemonDir = resolve(root, 'daemon')
mkdirSync(binDir, { recursive: true })

const out = resolve(binDir, 'aqnod' + ext)
let env = { ...process.env }
let buildArgs = '.'

if (useCgo) {
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
  console.log(`› building aqnod (pure Go) for ${triple}`)
}

execSync(`go build -o "${out}" ${buildArgs}`, { cwd: daemonDir, stdio: 'inherit', env })

const tripled = resolve(binDir, `aqnod-${triple}${ext}`)
cpSync(out, tripled)
console.log(`✓ sidecar ready → src-tauri/binaries/aqnod-${triple}${ext}`)
