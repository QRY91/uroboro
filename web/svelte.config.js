import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/vite-plugin-svelte').SvelteConfig} */
const config = {
  // Configure preprocessing for TypeScript
  preprocess: vitePreprocess(),

  compilerOptions: {
    // Enable source maps for debugging
    enableSourcemap: true,
  },

  // Suppress common warnings that aren't critical
  onwarn: (warning, handler) => {
    // Skip certain warnings during development
    if (warning.code === 'css-unused-selector') return;
    if (warning.code === 'a11y-click-events-have-key-events') return;
    if (warning.code === 'a11y-no-static-element-interactions') return;

    // Handle all other warnings normally
    handler(warning);
  },
};

export default config;
