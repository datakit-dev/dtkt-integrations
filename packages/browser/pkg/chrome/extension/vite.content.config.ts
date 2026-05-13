/**
 * Separate Vite build for the content script.
 *
 * Content scripts in Chrome MV3 must be classic (non-module) scripts —
 * the browser does not support `import` statements inside them. Setting
 * `format: 'iife'` bundles React and all other dependencies inline into a
 * single self-contained file with no top-level imports.
 *
 * This config runs after the main build (`vite build`) so that
 * `emptyOutDir: false` doesn't wipe the already-built popup/background assets.
 */
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  define: {
    // Inline NODE_ENV so React's dead-code elimination works correctly inside
    // the IIFE (there's no bundler runtime to provide process.env at runtime).
    'process.env.NODE_ENV': '"production"',
  },
  build: {
    outDir: path.resolve(__dirname, 'dist/assets'),
    emptyOutDir: false,
    rollupOptions: {
      input: {
        content: path.resolve(__dirname, 'src/content.ts'),
      },
      output: {
        format: 'iife',
        // Inline any dynamic imports rather than emitting separate chunks.
        inlineDynamicImports: true,
        entryFileNames: '[name].js',
        // Suppress the asset file-name hash so content.js lands at the path
        // declared in manifest.json.
        assetFileNames: '[name].[ext]',
      },
    },
  },
});
