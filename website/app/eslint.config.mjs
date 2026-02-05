import babelParser from '@babel/eslint-parser';
import importPlugin from 'eslint-plugin-import';
import vuePlugin from 'eslint-plugin-vue';
import vueParser from 'vue-eslint-parser';
import tsParser from '@typescript-eslint/parser';

export default [
  {
    files: ['**/*.{js,jsx,ts,tsx,vue,mjs,cjs}'],
    languageOptions: {
      parser: vueParser,
      parserOptions: {
        parser: babelParser,
        requireConfigFile: false,
        ecmaVersion: 2024,
        sourceType: 'module',
        ecmaFeatures: { jsx: true }
      },
      ecmaVersion: 2024,
      sourceType: 'module'
    },
    plugins: { vue: vuePlugin, import: importPlugin },
    settings: {
      'import/resolver': {
        alias: {
          map: [['@', './src']],
          extensions: ['.js', '.jsx', '.ts', '.tsx', '.vue', '.json']
        },
        node: { extensions: ['.js', '.jsx', '.ts', '.tsx', '.vue', '.json'] }
      }
    },
    rules: {
      'import/order': ['warn', {
        groups: ['builtin', 'external', 'internal', 'parent', 'sibling', 'index', 'object'],
        pathGroups: [
          { pattern: 'vue', group: 'builtin', position: 'before' },
          { pattern: '@/**', group: 'internal' }
        ],
        pathGroupsExcludedImportTypes: ['builtin'],
        'newlines-between': 'ignore',
        alphabetize: { order: 'ignore' }
      }],
      'vue/html-indent': ['error', 2],
      'no-unused-vars': ['warn', { argsIgnorePattern: '^_' }],
      'import/no-unresolved': 'error'
    }
  },
  {
    files: ['**/*.{ts,tsx,d.ts}'],
    languageOptions: {
      parser: tsParser,
      parserOptions: {
        ecmaVersion: 2024,
        sourceType: 'module'
      }
    },
    rules: {}
  }
];
