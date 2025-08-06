// OpenPenPal 统一ESLint配置 - SOTA级别的代码规范
// 设计原则：
// 1. 跨项目一致性：所有子项目共享相同的基础规则
// 2. 技术栈特化：React/Vue/Node.js特定的规则配置
// 3. 渐进式采用：可配置的严格程度，支持遗留代码
// 4. 自动修复：尽可能多的规则支持自动修复
// 5. 性能优化：缓存和增量检查支持

module.exports = {
  root: true,
  
  // 环境配置
  env: {
    browser: true,
    es2022: true,
    node: true,
  },
  
  // 扩展配置 - 分层继承
  extends: [
    'eslint:recommended',
    '@typescript-eslint/recommended',
    '@typescript-eslint/recommended-requiring-type-checking',
    'next/core-web-vitals',
    'plugin:import/recommended',
    'plugin:import/typescript',
    'plugin:promise/recommended',
    'plugin:security/recommended',
    'plugin:unicorn/recommended',
    'plugin:sonarjs/recommended',
    'prettier', // 必须在最后，覆盖格式化相关规则
  ],
  
  // 解析器配置
  parser: '@typescript-eslint/parser',
  parserOptions: {
    ecmaVersion: 2022,
    sourceType: 'module',
    ecmaFeatures: {
      jsx: true,
    },
    // TypeScript项目配置 - 支持多项目
    project: [
      './tsconfig.json',
      './frontend/tsconfig.json',
      './services/admin-service/frontend/tsconfig.json',
    ],
    tsconfigRootDir: __dirname,
  },
  
  // 插件配置
  plugins: [
    '@typescript-eslint',
    'react',
    'react-hooks',
    'import',
    'jsx-a11y',
    'promise',
    'security',
    'unicorn',
    'sonarjs',
  ],
  // SOTA级别规则配置
  rules: {
    // ========================
    // TypeScript规则增强
    // ========================
    '@typescript-eslint/no-unused-vars': [
      'error',
      {
        vars: 'all',
        args: 'after-used',
        ignoreRestSiblings: true,
        argsIgnorePattern: '^_',
        varsIgnorePattern: '^_',
      },
    ],
    '@typescript-eslint/no-explicit-any': [
      'warn',
      {
        fixToUnknown: true,
        ignoreRestArgs: true,
      },
    ],
    '@typescript-eslint/no-non-null-assertion': 'error',
    '@typescript-eslint/prefer-nullish-coalescing': 'error',
    '@typescript-eslint/prefer-optional-chain': 'error',
    '@typescript-eslint/consistent-type-imports': 'error',
    '@typescript-eslint/consistent-type-definitions': ['error', 'interface'],
    '@typescript-eslint/consistent-type-assertions': [
      'error',
      {
        assertionStyle: 'as',
        objectLiteralTypeAssertions: 'never',
      },
    ],
    '@typescript-eslint/await-thenable': 'error',
    '@typescript-eslint/no-floating-promises': 'error',
    '@typescript-eslint/no-misused-promises': 'error',

    // ========================
    // 命名规范
    // ========================
    '@typescript-eslint/naming-convention': [
      'error',
      {
        selector: 'variable',
        format: ['camelCase', 'PascalCase', 'UPPER_CASE'],
        leadingUnderscore: 'allow',
        trailingUnderscore: 'forbid',
      },
      {
        selector: 'function',
        format: ['camelCase'],
      },
      {
        selector: 'typeLike',
        format: ['PascalCase'],
      },
      {
        selector: 'interface',
        format: ['PascalCase'],
        custom: {
          regex: '^I[A-Z]',
          match: false,
        },
      },
    ],

    // ========================
    // React规则优化
    // ========================
    'react/jsx-uses-react': 'off',
    'react/react-in-jsx-scope': 'off',
    'react/prop-types': 'off',
    'react/display-name': 'error',
    'react/jsx-key': 'error',
    'react/jsx-no-duplicate-props': 'error',
    'react/jsx-no-undef': 'error',
    'react/jsx-uses-vars': 'error',
    'react/jsx-pascal-case': 'error',
    'react/no-danger-with-children': 'error',
    'react/no-deprecated': 'error',
    'react/no-direct-mutation-state': 'error',
    'react/no-find-dom-node': 'error',
    'react/no-is-mounted': 'error',
    'react/no-render-return-value': 'error',
    'react/no-string-refs': 'error',
    'react/no-unescaped-entities': 'error',
    'react/no-unknown-property': 'error',
    'react/require-render-return': 'error',

    // ========================
    // React Hooks增强
    // ========================
    'react-hooks/rules-of-hooks': 'error',
    'react-hooks/exhaustive-deps': 'warn',

    // ========================
    // 导入/导出规则增强
    // ========================
    'import/order': [
      'error',
      {
        groups: [
          'builtin',
          'external',
          'internal',
          ['parent', 'sibling'],
          'index',
          'object',
          'type',
        ],
        'newlines-between': 'always',
        alphabetize: {
          order: 'asc',
          caseInsensitive: true,
        },
        pathGroups: [
          {
            pattern: '@/**',
            group: 'internal',
            position: 'before',
          },
          {
            pattern: '~/**',
            group: 'internal',
            position: 'before',
          },
        ],
        pathGroupsExcludedImportTypes: ['builtin'],
      },
    ],
    'import/no-duplicates': 'error',
    'import/no-unresolved': 'error',
    'import/no-cycle': ['error', { maxDepth: 10 }],

    // ========================
    // 代码质量规则
    // ========================
    'no-console': process.env.NODE_ENV === 'production' ? 'error' : 'warn',
    'no-debugger': process.env.NODE_ENV === 'production' ? 'error' : 'warn',
    'no-alert': 'error',
    'no-var': 'error',
    'no-eval': 'error',
    'no-new-func': 'error',
    'no-implied-eval': 'error',
    'prefer-const': 'error',
    'prefer-template': 'error',
    'prefer-arrow-callback': 'error',
    'object-shorthand': 'error',
    'prefer-spread': 'error',
    'prefer-destructuring': [
      'error',
      {
        array: true,
        object: true,
      },
      {
        enforceForRenamedProperties: false,
      },
    ],

    // ========================
    // 代码复杂度控制
    // ========================
    complexity: ['warn', { max: 15 }],
    'max-depth': ['warn', { max: 4 }],
    'max-params': ['warn', { max: 5 }],
    'max-lines-per-function': [
      'warn',
      {
        max: 100,
        skipBlankLines: true,
        skipComments: true,
      },
    ],

    // ========================
    // 安全规则
    // ========================
    'security/detect-unsafe-regex': 'error',
    'security/detect-object-injection': 'error',

    // ========================
    // Unicorn规则调整
    // ========================
    'unicorn/no-null': 'off',
    'unicorn/no-array-reduce': 'off',
    'unicorn/no-array-for-each': 'off',
    'unicorn/filename-case': [
      'error',
      {
        cases: {
          camelCase: true,
          pascalCase: true,
          kebabCase: true,
        },
      },
    ],

    // ========================
    // SonarJS规则
    // ========================
    'sonarjs/cognitive-complexity': ['warn', 20],
    'sonarjs/no-duplicate-string': ['error', 5],
    'sonarjs/no-identical-conditions': 'error',

    // ========================
    // 可访问性规则保持
    // ========================
    'jsx-a11y/alt-text': 'error',
    'jsx-a11y/anchor-has-content': 'error',
    'jsx-a11y/anchor-is-valid': 'error',
    'jsx-a11y/aria-props': 'error',
    'jsx-a11y/aria-proptypes': 'error',
    'jsx-a11y/aria-unsupported-elements': 'error',
    'jsx-a11y/click-events-have-key-events': 'error',
    'jsx-a11y/heading-has-content': 'error',
    'jsx-a11y/img-redundant-alt': 'error',
    'jsx-a11y/interactive-supports-focus': 'error',
    'jsx-a11y/label-has-associated-control': 'error',
    'jsx-a11y/mouse-events-have-key-events': 'error',
    'jsx-a11y/no-access-key': 'error',
    'jsx-a11y/no-distracting-elements': 'error',
    'jsx-a11y/no-redundant-roles': 'error',
    'jsx-a11y/role-has-required-aria-props': 'error',
    'jsx-a11y/role-supports-aria-props': 'error',
    'jsx-a11y/scope': 'error',
  },
  settings: {
    react: {
      version: 'detect',
    },
    'import/resolver': {
      typescript: {
        alwaysTryTypes: true,
        project: './tsconfig.json',
      },
    },
  },
  // ========================
  // 覆盖配置 - 特定文件类型的规则
  // ========================
  
  overrides: [
    // Vue项目特定配置
    {
      files: ['**/admin-service/frontend/**/*.{js,ts,vue}'],
      extends: [
        'plugin:vue/vue3-recommended',
        '@vue/eslint-config-typescript',
      ],
      plugins: ['vue'],
      parser: 'vue-eslint-parser',
      parserOptions: {
        parser: '@typescript-eslint/parser',
      },
      rules: {
        // Vue特定规则
        'vue/multi-word-component-names': 'off',
        'vue/component-definition-name-casing': ['error', 'PascalCase'],
        'vue/component-name-in-template-casing': ['error', 'PascalCase'],
        'vue/prop-name-casing': ['error', 'camelCase'],
        'vue/attribute-hyphenation': ['error', 'always'],
        'vue/v-on-event-hyphenation': ['error', 'always'],
        
        // 禁用与Vue模板相关的TypeScript规则
        '@typescript-eslint/no-unsafe-assignment': 'off',
        '@typescript-eslint/no-unsafe-member-access': 'off',
      },
    },
    
    // Node.js服务特定配置
    {
      files: [
        'services/**/*.{js,ts}',
        'apps/**/*.{js,ts}',
        'scripts/**/*.{js,ts}',
        'backend/**/*.{js,ts}',
      ],
      excludedFiles: ['**/*.test.{js,ts}', '**/*.spec.{js,ts}'],
      env: {
        node: true,
        browser: false,
      },
      rules: {
        // Node.js特定规则
        'no-process-exit': 'error',
        'no-process-env': 'off', // 允许使用环境变量
        
        // 安全规则加强
        'security/detect-object-injection': 'error',
        'security/detect-buffer-noassert': 'error',
        'security/detect-child-process': 'warn',
        
        // 允许console.log在服务端
        'no-console': 'off',
        
        // 允许require
        '@typescript-eslint/no-var-requires': 'off',
      },
    },
    
    // 测试文件特定配置
    {
      files: ['**/*.{test,spec}.{js,ts,jsx,tsx}', '**/__tests__/**/*'],
      extends: ['plugin:jest/recommended'],
      plugins: ['jest'],
      env: {
        jest: true,
      },
      rules: {
        // 测试文件允许的规则
        '@typescript-eslint/no-explicit-any': 'off',
        '@typescript-eslint/no-non-null-assertion': 'off',
        'sonarjs/no-duplicate-string': 'off',
        'max-lines-per-function': 'off',
        'no-console': 'off',
        
        // Jest特定规则
        'jest/expect-expect': 'error',
        'jest/no-disabled-tests': 'warn',
        'jest/no-focused-tests': 'error',
        'jest/prefer-to-have-length': 'warn',
        'jest/valid-expect': 'error',
      },
    },
    
    // 配置文件特殊规则
    {
      files: [
        '*.config.{js,ts}',
        '**/*.config.{js,ts}',
        '**/webpack.config.{js,ts}',
        '**/vite.config.{js,ts}',
        '**/rollup.config.{js,ts}',
        '**/next.config.{js,ts}',
        '**/tailwind.config.{js,ts}',
        '**/postcss.config.{js,ts}',
        '**/scripts/**/*',
      ],
      rules: {
        // 配置文件允许使用require
        '@typescript-eslint/no-var-requires': 'off',
        'import/no-extraneous-dependencies': 'off',
        'unicorn/prefer-module': 'off',
        'no-console': 'off',
      },
    },
    
    // Python文件（忽略JS规则）
    {
      files: ['**/*.py'],
      parser: 'espree',
      rules: {},
    },
    
    // Go文件（忽略JS规则）
    {
      files: ['**/*.go'],
      parser: 'espree', 
      rules: {},
    },
    
    // Shell脚本
    {
      files: ['**/*.{sh,bash}'],
      parser: 'espree',
      rules: {},
    },
    
    // Markdown文件中的代码块
    {
      files: ['**/*.md'],
      parser: 'eslint-plugin-markdown/parser',
      rules: {
        'no-console': 'off',
        '@typescript-eslint/no-unused-vars': 'off',
        'import/no-unresolved': 'off',
      },
    },
  ],
  
  // ========================
  // 忽略配置
  // ========================
  
  ignorePatterns: [
    'node_modules/',
    'dist/',
    'build/',
    'coverage/',
    '.next/',
    '.nuxt/',
    '.cache/',
    'public/',
    '*.min.js',
    '*.bundle.js',
    'vendor/',
    'lib/',
    'types/generated/',
    '**/*.d.ts',
    // 临时忽略遗留代码目录（逐步迁移）
    'legacy/',
    'old/',
    'archive/',
    'backup/',
    'temp/',
    'tmp/',
    // 忽略生成的文件
    'venv/',
    '__pycache__/',
    '*.pyc',
    'target/',
    'bin/',
    // 忽略日志文件
    'logs/',
    '*.log',
  ],
};