import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';

export default defineConfig({
  plugins: [svelte()],
  build: {
    // WebView2 is evergreen Chromium: no legacy transforms needed.
    target: 'chrome110',
    reportCompressedSize: false,
  },
});
