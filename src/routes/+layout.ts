// SPA mode: the Tauri webview renders everything client-side. No SSR, no
// prerender — adapter-static serves an index.html fallback and we route here.
export const ssr = false
export const prerender = false
export const trailingSlash = 'never'
