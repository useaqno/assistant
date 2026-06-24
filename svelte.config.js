import adapter from '@sveltejs/adapter-static'
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte'

/** @type {import('@sveltejs/kit').Config} */
const config = {
  preprocess: vitePreprocess(),
  kit: {
    // Tauri serves a static bundle and we route on the client (SPA), so we use
    // adapter-static with an index.html fallback.
    adapter: adapter({
      fallback: 'index.html',
      precompress: false,
      strict: false
    }),
    alias: {
      $components: 'src/lib/components',
      $stores: 'src/lib/stores'
    }
  }
}

export default config
