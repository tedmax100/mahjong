import { defineConfig } from 'vite';
import path from 'path';

export default defineConfig({
  server: {
    host: true,
    port: 5174,
    allowedHosts: ['.trycloudflare.com'],
    proxy: {
      '/api/lobby': {
        target: 'http://localhost:3001',
        changeOrigin: true,
      },
      '/ws/lobby': {
        target: 'ws://localhost:3001',
        ws: true,
      },
      '/login': {
        target: 'http://localhost:3001',
        changeOrigin: true,
      },
      '/auth': {
        target: 'http://localhost:3001',
        changeOrigin: true,
      },
    },
  },
  envDir: '.',
  resolve: {
    alias: {
      '@shared': path.resolve(__dirname, '../shared'),
    },
  },
  build: {
    outDir: '../dist/lobby',
    emptyOutDir: true,
  },
});
