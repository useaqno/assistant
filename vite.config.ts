import { sveltekit } from '@sveltejs/vite-plugin-svelte';
import { defineConfig } from 'vite';

// Tauri expects a fixed dev server on 1420/5173 and must not clobber its own
// console output. See https://tauri.app for details.
const host = process.env.TAURI_DEV_HOST;

export default defineConfig({
  plugins: [sveltekit()],
  clearScreen: false,
  server: {
    port: 5173,
    strictPort: true,
    host: host || false,
    hmr: host ? { protocol: 'ws', host, port: 5174 } : undefined,
    watch: { ignored: ['**/src-tauri/**', '**/daemon/**'] }
  },
  // Env vars starting with these prefixes are exposed to the client.
  envPrefix: ['VITE_', 'TAURI_']
});
