import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// The SPA is served from /admin/ by the Go server and embedded via go:embed.
// https://vite.dev/config/
export default defineConfig({
  base: '/admin/',
  plugins: [react()],
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
})
