// OpenPenPal 统一Prettier配置 - SOTA级别的代码格式化
// 设计原则：
// 1. 与ESLint配合：避免规则冲突，专注格式化
// 2. 跨项目一致性：所有子项目使用相同的格式化规则
// 3. 可读性优先：优化代码可读性和维护性
// 4. 团队协作：减少代码格式争议

module.exports = {
  // ========================
  // 基础格式化配置
  // ========================
  
  // 行宽限制 - 平衡可读性和屏幕空间利用
  printWidth: 100,
  
  // 缩进配置
  tabWidth: 2,
  useTabs: false, // 使用空格而不是制表符
  
  // 分号配置
  semi: true, // 总是使用分号，避免ASI陷阱
  
  // 引号配置  
  singleQuote: true, // 使用单引号，与ESLint保持一致
  quoteProps: 'as-needed', // 仅在需要时为对象属性添加引号
  
  // JSX配置
  jsxSingleQuote: true, // JSX中也使用单引号
  bracketSameLine: false, // JSX标签的>换行
  
  // 尾随逗号配置
  trailingComma: 'all', // 总是使用尾随逗号，有利于版本控制
  
  // 括号间距
  bracketSpacing: true, // 对象字面量括号内添加空格
  
  // 箭头函数参数括号
  arrowParens: 'avoid', // 单参数箭头函数不使用括号
  
  // ========================
  // 换行和空白配置
  // ========================
  
  // 换行符配置
  endOfLine: 'lf', // 统一使用LF换行符，兼容不同操作系统
  
  // HTML空白敏感性
  htmlWhitespaceSensitivity: 'css',
  
  // Vue文件中的脚本和样式标签缩进
  vueIndentScriptAndStyle: true,
  
  // 其他配置
  insertPragma: false,
  requirePragma: false,
  proseWrap: 'preserve',
  
  // ========================
  // 文件特定覆盖配置
  // ========================
  
  overrides: [
    // TypeScript文件
    {
      files: '*.{ts,tsx}',
      options: {
        parser: 'typescript',
        semi: true,
        singleQuote: true,
        trailingComma: 'all',
      },
    },
    
    // JavaScript文件
    {
      files: '*.{js,jsx}',
      options: {
        parser: 'babel',
        semi: true,
        singleQuote: true,
        trailingComma: 'all',
      },
    },
    
    // Vue单文件组件
    {
      files: '*.vue',
      options: {
        parser: 'vue',
        vueIndentScriptAndStyle: true,
        htmlWhitespaceSensitivity: 'ignore',
        singleQuote: true,
      },
    },
    
    // JSON文件
    {
      files: '*.json',
      options: {
        parser: 'json',
        printWidth: 120, // JSON可以使用更宽的行宽
        trailingComma: 'none', // JSON不使用尾随逗号
        singleQuote: false, // JSON使用双引号
        tabWidth: 2,
      },
    },
    
    // YAML文件
    {
      files: ['*.yml', '*.yaml'],
      options: {
        parser: 'yaml',
        tabWidth: 2,
        singleQuote: false, // YAML通常使用双引号
        printWidth: 100,
      },
    },
    
    // Markdown文件
    {
      files: '*.{md,mdx}',
      options: {
        parser: 'markdown',
        printWidth: 100, // Markdown使用标准行宽
        proseWrap: 'always', // 总是换行，便于版本控制
        embeddedLanguageFormatting: 'auto', // 格式化内嵌代码
      },
    },
    
    // CSS/SCSS/Less文件
    {
      files: '*.{css,scss,sass,less}',
      options: {
        parser: 'css',
        singleQuote: true,
        printWidth: 100,
      },
    },
    
    // HTML文件
    {
      files: '*.html',
      options: {
        parser: 'html',
        printWidth: 120, // HTML可以使用更长的行宽
        htmlWhitespaceSensitivity: 'ignore',
        singleAttributePerLine: false,
      },
    },
    
    // 配置文件特殊处理
    {
      files: [
        '.eslintrc.js',
        '.prettierrc.js',
        '*.config.js',
        '*.config.ts',
        'next.config.js',
        'tailwind.config.js',
        'vite.config.ts',
      ],
      options: {
        printWidth: 120, // 配置文件可以使用更长的行宽
        bracketSpacing: true,
        trailingComma: 'all',
      },
    },
    
    // package.json特殊处理
    {
      files: 'package.json',
      options: {
        parser: 'json-stringify',
        printWidth: 120,
        tabWidth: 2,
        trailingComma: 'none',
        singleQuote: false,
      },
    },
    
    // README和文档文件
    {
      files: ['README.md', 'CHANGELOG.md', 'CONTRIBUTING.md', 'docs/**/*.md'],
      options: {
        printWidth: 100,
        proseWrap: 'always',
        embeddedLanguageFormatting: 'auto', // 格式化代码块
      },
    },
    
    // SQL文件
    {
      files: '*.sql',
      options: {
        parser: 'sql',
        printWidth: 120,
        tabWidth: 2,
        useTabs: false,
      },
    },
    
    // GraphQL文件
    {
      files: '*.{gql,graphql}',
      options: {
        parser: 'graphql',
        printWidth: 120,
      },
    },
  ],
};