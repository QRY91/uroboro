module.exports = {
  root: true,
  env: {
    browser: true,
    es2022: true,
    node: true
  },
  extends: [
    'eslint:recommended',
    '@typescript-eslint/recommended',
    'plugin:svelte/recommended',
    'plugin:@typescript-eslint/recommended-requiring-type-checking'
  ],
  parser: '@typescript-eslint/parser',
  plugins: ['@typescript-eslint'],
  parserOptions: {
    ecmaVersion: 2022,
    sourceType: 'module',
    tsconfigRootDir: __dirname,
    project: ['./tsconfig.json'],
    extraFileExtensions: ['.svelte']
  },
  overrides: [
    {
      files: ['*.svelte'],
      parser: 'svelte-eslint-parser',
      parserOptions: {
        parser: '@typescript-eslint/parser'
      }
    },
    {
      files: ['*.js', '*.cjs'],
      extends: ['eslint:recommended'],
      rules: {
        '@typescript-eslint/no-var-requires': 'off'
      }
    }
  ],
  rules: {
    // TypeScript specific rules
    '@typescript-eslint/no-unused-vars': ['error', { argsIgnorePattern: '^_' }],
    '@typescript-eslint/no-explicit-any': 'warn',
    '@typescript-eslint/explicit-function-return-type': 'off',
    '@typescript-eslint/explicit-module-boundary-types': 'off',
    '@typescript-eslint/no-non-null-assertion': 'warn',
    '@typescript-eslint/prefer-const': 'error',
    '@typescript-eslint/no-unnecessary-type-assertion': 'error',

    // General code quality
    'no-console': 'warn',
    'no-debugger': 'error',
    'no-alert': 'error',
    'prefer-const': 'error',
    'no-var': 'error',
    'object-shorthand': 'error',
    'prefer-template': 'error',

    // Import/Export rules
    'no-duplicate-imports': 'error',

    // Svelte specific rules
    'svelte/no-at-debug-tags': 'warn',
    'svelte/no-target-blank': 'error',
    'svelte/no-at-html-tags': 'warn',
    'svelte/html-self-closing': ['error', 'all'],
    'svelte/mustache-spacing': 'error',
    'svelte/no-spaces-around-equal-signs-in-attribute': 'error',
    'svelte/prefer-class-directive': 'error',
    'svelte/prefer-style-directive': 'error',
    'svelte/shorthand-attribute': 'error',
    'svelte/shorthand-directive': 'error',
    'svelte/sort-attributes': 'warn',
    'svelte/spaced-html-comment': 'error',

    // Accessibility rules
    'svelte/a11y-alt-text': 'error',
    'svelte/a11y-aria-attributes': 'error',
    'svelte/a11y-click-events-have-key-events': 'error',
    'svelte/a11y-img-redundant-alt': 'error',
    'svelte/a11y-label-has-associated-control': 'error',
    'svelte/a11y-media-has-caption': 'error',
    'svelte/a11y-missing-attribute': 'error',
    'svelte/a11y-mouse-events-have-key-events': 'error',
    'svelte/a11y-no-redundant-roles': 'error',
    'svelte/a11y-role-has-required-aria-props': 'error',
    'svelte/a11y-structure': 'error',
    'svelte/a11y-tabindex-no-positive': 'error',

    // Performance and best practices
    'svelte/no-reactive-functions': 'error',
    'svelte/no-reactive-literals': 'error',
    'svelte/no-useless-mustaches': 'error',
    'svelte/prefer-destructuring-props': 'error',
    'svelte/require-stores-init': 'error',
    'svelte/valid-compile': 'error',
    'svelte/valid-prop-names-in-kit-pages': 'error'
  },
  settings: {
    'svelte3/typescript': true
  },
  ignorePatterns: [
    'dist/',
    'build/',
    'node_modules/',
    '*.cjs',
    'vite.config.ts'
  ]
};
