// Builds the Go daemon and names it the way Tauri expects a sidecar:
//   src-tauri/binaries/aqnod-<rust-host-target-triple>[.exe]
//
// Run via `pnpm daemon:sidecar`. The `pnpm app:dev` script runs this first so
// `tauri dev` always finds an up-to-date binary.
import { execSync } from 'node:child_process';
import { cpSync, mkdirSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, resolve } from 'node:path';

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..');
const ext = process.platform === 'win32' ? '.exe' : '';

function hostTriple() {
  const out = execSync('rustc -Vv', { encoding: 'utf8' });
  const line = out.split('\n').find((l) => l.startsWith('host:'));
  if (!line) throw new Error('Could not determine rust host triple — is rustc installed?');
  return line.split(':')[1].trim();
}

const triple = hostTriple();
const binDir = resolve(root, 'src-tauri/binaries');
mkdirSync(binDir, { recursive: true });

console.log(`› building aqnod for ${triple}`);
execSync(`go build -o "${resolve(binDir, 'aqnod' + ext)}" .`, {
  cwd: resolve(root, 'daemon'),
  stdio: 'inherit'
});

const tripled = resolve(binDir, `aqnod-${triple}${ext}`);
cpSync(resolve(binDir, `aqnod${ext}`), tripled);
console.log(`✓ sidecar ready → src-tauri/binaries/aqnod-${triple}${ext}`);
