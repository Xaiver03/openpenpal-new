module.exports = {
  extends: [
    'next/core-web-vitals',
    'plugin:@typescript-eslint/recommended'
  ],
  parser: '@typescript-eslint/parser',
  parserOptions: {
    ecmaVersion: 2022,
    sourceType: 'module',
    project: './tsconfig.json',
  },
  ignorePatterns: ['*.config.js', '.eslintrc.js'],
  rules: {
    // Allow any for now
    '@typescript-eslint/no-explicit-any': 'off',
    // Allow non-null assertions
    '@typescript-eslint/no-non-null-assertion': 'off',
    // Allow unused vars with underscore prefix
    '@typescript-eslint/no-unused-vars': [
      'warn',
      {
        argsIgnorePattern: '^_',
        varsIgnorePattern: '^_',
      },
    ],
    // Allow require
    '@typescript-eslint/no-var-requires': 'off',
    // Disable some strict checks temporarily
    '@typescript-eslint/no-unsafe-assignment': 'off',
    '@typescript-eslint/no-unsafe-member-access': 'off',
    '@typescript-eslint/no-unsafe-call': 'off',
    '@typescript-eslint/no-unsafe-return': 'off',
  },
}