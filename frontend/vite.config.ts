import react from '@vitejs/plugin-react';
import { defineConfig } from 'vite';

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    host: '0.0.0.0',
    allowedHosts: ['localhost', 'frontend'],
    proxy: {
      // Auth API
      '/api/auth': {
        target: 'http://localhost:8001',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/auth/, ''),
      },
      // Boards API
      '/api/boards': {
        target: 'http://localhost:8002',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/boards/, ''),
      },
      // Realtime WS
      '/api/realtime': {
        target: 'http://localhost:8003',
        changeOrigin: true,
        ws: true,
        rewrite: (path) => path.replace(/^\/api\/realtime/, ''),
      },
    },
  },
  build: {
    outDir: 'dist',
  },
});


