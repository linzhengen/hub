import js from '@eslint/js';
import reactHooks from 'eslint-plugin-react-hooks';
import reactRefresh from 'eslint-plugin-react-refresh';
import globals from 'globals';
import tseslint from 'typescript-eslint';

export default tseslint.config(
  { ignores: ['dist', 'node_modules'] },
  {
    files: ['**/*.{ts,tsx}'],
    extends: [js.configs.recommended, ...tseslint.configs.recommended],
    languageOptions: {
      ecmaVersion: 2022,
      globals: globals.browser,
    },
    plugins: {
      'react-hooks': reactHooks,
      'react-refresh': reactRefresh,
    },
    rules: {
      // The recommended preset bundles the React Compiler-oriented rules
      // (immutability/purity/set-state-in-effect/etc.) as errors. They were
      // all downgraded to warnings while the codebase still violated them;
      // the violations are gone, so the preset's levels now apply as-is.
      ...reactHooks.configs.recommended.rules,
      // The one exception: AppSidebar derives its open submenu from the
      // current route inside an effect. Rewriting that means deciding how a
      // manual toggle should interact with navigation, which is a behaviour
      // change rather than a refactor. Tracked in #67.
      'react-hooks/set-state-in-effect': 'warn',
      'react-refresh/only-export-components': [
        'warn',
        { allowConstantExport: true },
      ],
      '@typescript-eslint/no-explicit-any': 'error',
      '@typescript-eslint/no-unused-vars': [
        'error',
        { argsIgnorePattern: '^_', varsIgnorePattern: '^_' },
      ],
    },
  },
);
