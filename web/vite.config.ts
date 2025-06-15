import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';
import { resolve } from 'path';

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    svelte({
      preprocess: vitePreprocess(),
      compilerOptions: {
        dev: process.env.NODE_ENV === 'development',
      },
    }),
  ],

  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
      '@lib': resolve(__dirname, 'src/lib'),
      '@components': resolve(__dirname, 'src/components'),
      '@stores': resolve(__dirname, 'src/stores'),
      '@types': resolve(__dirname, 'src/types'),
      '@utils': resolve(__dirname, 'src/utils'),
    },
  },

  build: {
    target: 'es2018', // Support for slightly older devices
    minify: true,
    cssMinify: true,
    rollupOptions: {
      output: {
        manualChunks: {
          // Split vendor chunks for better caching
          vendor: ['svelte'],
          animations: ['animejs'],
        },
      },
    },
    // Optimize for low-spec devices
    chunkSizeWarningLimit: 1000,
    sourcemap: false, // Disable in production for smaller bundle
  },

  server: {
    port: 3000,
    host: true, // Allow external connections
    proxy: {
      // Proxy API calls to Go backend
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        secure: false,
      },
      '/journey': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        secure: false,
      },
    },
  },

  preview: {
    port: 3000,
    host: true,
  },

  optimizeDeps: {
    include: ['animejs', 'd3-scale', 'd3-time', 'd3-time-format'],
    exclude: ['lucide-svelte'],
  },

  css: {
    preprocessorOptions: {
      scss: {
        additionalData: `@import "@/styles/variables.scss";`,
      },
    },
  },
});
