import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vitest/config'

// Kept separate from vite.config.ts on purpose: Vitest resolves its own
// Vite, so sharing one config file makes `vue-tsc -b` fail on a conflict
// between two incompatible `Plugin` types. This file is excluded from the
// tsconfig project for the same reason — Vitest transpiles it itself.
export default defineConfig({
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  plugins: [vue() as any],
  test: {
    // jsdom supplies `window` and the `storage` event. It only exposes
    // `localStorage` for a real origin — the default `about:blank` is an
    // opaque origin, where the Storage API is unavailable — and Node 26
    // additionally shadows it with an experimental global that is inert
    // without `--localstorage-file`. src/test-setup.ts settles both.
    environment: 'jsdom',
    environmentOptions: { jsdom: { url: 'http://localhost:5173' } },
    setupFiles: ['src/test-setup.ts'],
    include: ['src/**/*.test.ts'],
  },
})
