/**
 * ESLint 配置
 * 使用 TypeScript 和 Vue 3 最佳实践
 */

module.exports = {
  root: true,
  env: {
    browser: true,
    es2021: true,
    node: true,
  },
  extends: [
    'eslint:recommended',
    'plugin:@typescript-eslint/recommended',
    'plugin:vue/vue3-recommended',
  ],
  parserOptions: {
    parser: '@typescript-eslint/parser',
    ecmaVersion: 'latest',
    sourceType: 'module',
  },
  rules: {
    // Vue 相关
    'vue/multi-word-component-names': 'off',
    'vue/max-attributes-per-line': ['error', { singleline: 3, multiline: 1 }],
    'vue/component-definition-name-casing': ['error', 'PascalCase'],

    // TypeScript 相关
    '@typescript-eslint/no-unused-vars': ['error', { argsIgnorePattern: '^_' }],
    '@typescript-eslint/explicit-function-return-types': 'off',
    '@typescript-eslint/no-explicit-any': 'warn',

    // 通用规则
    'no-console': ['warn', { allow: ['warn', 'error'] }],
    semi: ['error', 'never'],
    quotes: ['error', 'single', { avoidEscape: true }],
    'comma-dangle': ['error', 'always-multiline'],
    'no-trailing-spaces': 'error',
    'eol-last': ['error', 'always'],
  },
}
