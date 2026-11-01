import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vite'

// https://vite.dev/config/
// Test configuration lives in vitest.config.ts: Vitest bundles its own copy
// of Vite, and merging the two plugin type universes in one file breaks the
// production type-check.
export default defineConfig({
  plugins: [vue()],
})
