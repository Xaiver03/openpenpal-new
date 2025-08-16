/**
 * ESLint Configuration for Circular Dependency Prevention
 * ESLint 循环依赖预防配置
 * 
 * Purpose: Add ESLint rules to detect and prevent circular dependencies
 * 目的: 添加 ESLint 规则来检测和预防循环依赖
 */

module.exports = {
  extends: [
    'next/core-web-vitals'
  ],
  plugins: [
    'import',
    '@typescript-eslint'
  ],
  rules: {
    // ================================
    // Circular Dependency Prevention
    // ================================
    
    /**
     * Detect circular dependencies
     * 检测循环依赖
     */
    'import/no-cycle': ['error', {
      // Maximum depth to traverse when looking for cycles
      maxDepth: 10,
      
      // Ignore external modules (node_modules)
      ignoreExternal: true,
      
      // Allow imports that form a cycle, but show warnings
      // Set to 'error' to strictly prevent cycles
      severity: 'error'
    }],
    
    /**
     * Prevent self-imports (importing from the same file)
     * 防止自导入（从同一文件导入）
     */
    'import/no-self-import': 'error',
    
    /**
     * Ensure imports are properly ordered and grouped
     * 确保导入按正确顺序分组
     */
    'import/order': ['warn', {
      'groups': [
        'builtin',     // Node.js built-ins
        'external',    // npm packages
        'internal',    // Internal modules (absolute imports)
        'parent',      // Parent directory imports
        'sibling',     // Same directory imports
        'index',       // Index file imports
        'type'         // Type-only imports
      ],
      'newlines-between': 'always',
      'pathGroups': [
        {
          'pattern': '@/**',
          'group': 'internal',
          'position': 'before'
        },
        {
          'pattern': '@/lib/**',
          'group': 'internal',
          'position': 'before'
        },
        {
          'pattern': '@/stores/**',
          'group': 'internal',
          'position': 'after'
        },
        {
          'pattern': '@/contexts/**',
          'group': 'internal',
          'position': 'after'
        }
      ],
      'pathGroupsExcludedImportTypes': ['type'],
      'alphabetize': {
        'order': 'asc',
        'caseInsensitive': true
      }
    }],
    
    /**
     * Warn about unused imports (can indicate refactoring opportunities)
     * 警告未使用的导入（可能表明重构机会）
     */
    'import/no-unused-modules': ['warn', {
      'unusedExports': true,
      'missingExports': false
    }],
    
    /**
     * Prevent importing from paths that might create cycles
     * 防止从可能创建循环的路径导入
     */
    'import/no-restricted-paths': ['error', {
      'zones': [
        {
          // Services should not import from contexts or stores
          'target': './src/lib/services/**',
          'from': './src/contexts/**',
          'message': 'Services should not directly import contexts. Use dependency injection instead.'
        },
        {
          'target': './src/lib/services/**',
          'from': './src/stores/**',
          'message': 'Services should not directly import stores. Use dependency injection instead.'
        },
        {
          // Contexts should not import from each other
          'target': './src/contexts/**',
          'from': './src/contexts/**',
          'except': ['./src/contexts/types.ts'],
          'message': 'Contexts should not import from other contexts. Use shared types file instead.'
        },
        {
          // Stores should not import contexts directly
          'target': './src/stores/**',
          'from': './src/contexts/**',
          'message': 'Stores should not directly import contexts. Use dependency injection instead.'
        },
        {
          // API modules should not import services (except API client)
          'target': './src/lib/api/**',
          'from': './src/lib/services/**',
          'except': ['./src/lib/services/service-factory.ts'],
          'message': 'API modules should not import services directly. Use service factory instead.'
        }
      ]
    }],
    
    // ================================
    // TypeScript Specific Rules
    // ================================
    
    /**
     * Prefer type-only imports when possible
     * 尽可能使用仅类型导入
     */
    '@typescript-eslint/consistent-type-imports': ['warn', {
      'prefer': 'type-imports',
      'disallowTypeAnnotations': false,
      'fixStyle': 'separate-type-imports'
    }],
    
    /**
     * Require explicit return types on exported functions
     * 要求导出函数有明确的返回类型
     */
    '@typescript-eslint/explicit-module-boundary-types': 'warn',
    
    // ================================
    // React Specific Rules  
    // ================================
    
    /**
     * Prevent components from importing hooks that might cause cycles
     * 防止组件导入可能导致循环的hooks
     */
    'react-hooks/rules-of-hooks': 'error',
    'react-hooks/exhaustive-deps': 'warn'
  },
  
  // ================================
  // Environment-specific Overrides
  // ================================
  
  overrides: [
    {
      // More strict rules for service layer
      files: ['src/lib/services/**/*.ts', 'src/lib/services/**/*.tsx'],
      rules: {
        'import/no-cycle': ['error', { maxDepth: 5 }],
        '@typescript-eslint/explicit-module-boundary-types': 'error'
      }
    },
    {
      // Special rules for DI layer
      files: ['src/lib/di/**/*.ts'],
      rules: {
        'import/no-cycle': ['error', { maxDepth: 3 }],
        'import/no-self-import': 'error'
      }
    },
    {
      // Allow more flexibility in test files
      files: ['**/*.test.ts', '**/*.test.tsx', '**/*.spec.ts', '**/*.spec.tsx'],
      rules: {
        'import/no-cycle': 'warn',
        'import/no-restricted-paths': 'off'
      }
    },
    {
      // Special rules for index files (barrel exports)
      files: ['**/index.ts', '**/index.tsx'],
      rules: {
        'import/no-cycle': ['error', { maxDepth: 8 }],
        'import/no-unused-modules': 'off'
      }
    }
  ],
  
  // ================================
  // Parser and Settings
  // ================================
  
  parser: '@typescript-eslint/parser',
  parserOptions: {
    ecmaVersion: 2022,
    sourceType: 'module',
    ecmaFeatures: {
      jsx: true
    },
    project: './tsconfig.json'
  },
  
  settings: {
    'import/resolver': {
      'typescript': {
        'project': './tsconfig.json',
        'alwaysTryTypes': true
      },
      'node': {
        'extensions': ['.js', '.jsx', '.ts', '.tsx']
      }
    },
    'import/parsers': {
      '@typescript-eslint/parser': ['.ts', '.tsx']
    },
    'import/extensions': ['.js', '.jsx', '.ts', '.tsx']
  }
}