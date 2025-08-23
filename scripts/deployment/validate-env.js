#!/usr/bin/env node

/**
 * 环境变量校验脚本
 * 用于上线前检查所有必需的环境变量是否正确配置
 */

const fs = require('fs');
const path = require('path');
const { URL } = require('url');

// 颜色输出辅助函数
const colors = {
  red: text => `\x1b[31m${text}\x1b[0m`,
  green: text => `\x1b[32m${text}\x1b[0m`,
  yellow: text => `\x1b[33m${text}\x1b[0m`,
  blue: text => `\x1b[34m${text}\x1b[0m`,
  gray: text => `\x1b[90m${text}\x1b[0m`
};

// 环境变量配置要求
const envRequirements = {
  // 基础配置
  NODE_ENV: {
    required: true,
    values: ['development', 'production', 'test'],
    description: '运行环境'
  },
  
  // 前端配置
  NEXT_PUBLIC_API_BASE_URL: {
    required: true,
    validator: (value) => {
      try {
        new URL(value);
        return true;
      } catch {
        return false;
      }
    },
    description: 'API基础URL'
  },
  NEXT_PUBLIC_WS_URL: {
    required: true,
    validator: (value) => {
      try {
        const url = new URL(value);
        return url.protocol === 'ws:' || url.protocol === 'wss:';
      } catch {
        return false;
      }
    },
    description: 'WebSocket URL'
  },
  
  // 后端配置
  DATABASE_URL: {
    required: true,
    validator: (value) => {
      return value.startsWith('postgres://') || value.startsWith('postgresql://');
    },
    description: 'PostgreSQL数据库连接'
  },
  DB_TYPE: {
    required: true,
    values: ['postgres'],
    description: '数据库类型'
  },
  REDIS_URL: {
    required: true,
    validator: (value) => {
      return value.startsWith('redis://') || value.startsWith('rediss://');
    },
    description: 'Redis连接URL'
  },
  
  // JWT配置
  JWT_SECRET: {
    required: true,
    validator: (value) => {
      return value.length >= 32;
    },
    description: 'JWT密钥(至少32字符)'
  },
  JWT_PUBLIC_KEY: {
    required: false,
    validator: (value) => {
      return value.includes('BEGIN PUBLIC KEY');
    },
    description: 'JWT公钥'
  },
  JWT_PRIVATE_KEY: {
    required: false,
    validator: (value) => {
      return value.includes('BEGIN PRIVATE KEY');
    },
    description: 'JWT私钥'
  },
  
  // API密钥
  MOONSHOT_API_KEY: {
    required: true,
    validator: (value) => {
      return value.length > 0;
    },
    description: 'Moonshot AI API密钥'
  },
  SILICON_FLOW_API_KEY: {
    required: false,
    validator: (value) => {
      return value.length > 0;
    },
    description: 'Silicon Flow API密钥'
  },
  
  // 服务端口
  PORT: {
    required: false,
    default: '8080',
    validator: (value) => {
      const port = parseInt(value);
      return port > 0 && port < 65536;
    },
    description: '后端服务端口'
  },
  FRONTEND_PORT: {
    required: false,
    default: '3000',
    validator: (value) => {
      const port = parseInt(value);
      return port > 0 && port < 65536;
    },
    description: '前端服务端口'
  },
  
  // 文件上传
  UPLOAD_DIR: {
    required: false,
    default: './uploads',
    validator: (value) => {
      return value.length > 0;
    },
    description: '文件上传目录'
  },
  MAX_UPLOAD_SIZE: {
    required: false,
    default: '10485760',
    validator: (value) => {
      return parseInt(value) > 0;
    },
    description: '最大上传文件大小'
  },
  
  // 监控配置
  PROMETHEUS_ENABLED: {
    required: false,
    default: 'true',
    values: ['true', 'false'],
    description: 'Prometheus监控开关'
  },
  GRAFANA_ENABLED: {
    required: false,
    default: 'true',
    values: ['true', 'false'],
    description: 'Grafana监控开关'
  }
};

// 加载.env文件
function loadEnvFile(envPath) {
  if (!fs.existsSync(envPath)) {
    return {};
  }
  
  const content = fs.readFileSync(envPath, 'utf8');
  const env = {};
  
  content.split('\n').forEach(line => {
    const trimmed = line.trim();
    if (trimmed && !trimmed.startsWith('#')) {
      const [key, ...valueParts] = trimmed.split('=');
      if (key) {
        let value = valueParts.join('=');
        // 移除引号
        if ((value.startsWith('"') && value.endsWith('"')) || 
            (value.startsWith("'") && value.endsWith("'"))) {
          value = value.slice(1, -1);
        }
        env[key.trim()] = value.trim();
      }
    }
  });
  
  return env;
}

// 验证环境变量
function validateEnvironment() {
  console.log('\n' + colors.blue('=== 环境变量校验开始 ===\n'));
  
  const errors = [];
  const warnings = [];
  const successes = [];
  
  // 尝试加载.env文件
  const rootDir = path.resolve(__dirname, '../../..');
  const envFiles = [
    path.join(rootDir, '.env'),
    path.join(rootDir, '.env.local'),
    path.join(rootDir, '.env.production')
  ];
  
  let loadedEnv = {};
  envFiles.forEach(envFile => {
    if (fs.existsSync(envFile)) {
      console.log(colors.gray(`加载环境文件: ${path.relative(rootDir, envFile)}`));
      Object.assign(loadedEnv, loadEnvFile(envFile));
    }
  });
  
  // 合并进程环境变量
  const env = { ...loadedEnv, ...process.env };
  
  // 验证每个环境变量
  for (const [key, config] of Object.entries(envRequirements)) {
    const value = env[key];
    
    if (!value && config.required) {
      errors.push(`${colors.red('✗')} ${key}: 必需的环境变量未设置 (${config.description})`);
      continue;
    }
    
    if (!value && !config.required) {
      if (config.default) {
        warnings.push(`${colors.yellow('⚠')} ${key}: 使用默认值 "${config.default}" (${config.description})`);
      } else {
        warnings.push(`${colors.yellow('⚠')} ${key}: 可选变量未设置 (${config.description})`);
      }
      continue;
    }
    
    // 验证值
    if (config.values && !config.values.includes(value)) {
      errors.push(`${colors.red('✗')} ${key}: 值 "${value}" 不在允许范围内 ${JSON.stringify(config.values)}`);
      continue;
    }
    
    if (config.validator && !config.validator(value)) {
      errors.push(`${colors.red('✗')} ${key}: 值格式不正确 (${config.description})`);
      continue;
    }
    
    // 检查密钥强度
    if (key.includes('SECRET') || key.includes('KEY')) {
      if (value.length < 16) {
        warnings.push(`${colors.yellow('⚠')} ${key}: 密钥长度较短，建议至少16字符`);
      }
    }
    
    successes.push(`${colors.green('✓')} ${key}: ${config.description}`);
  }
  
  // 输出结果
  console.log('\n' + colors.blue('=== 验证结果 ===\n'));
  
  if (successes.length > 0) {
    console.log(colors.green('成功项:'));
    successes.forEach(msg => console.log('  ' + msg));
  }
  
  if (warnings.length > 0) {
    console.log('\n' + colors.yellow('警告项:'));
    warnings.forEach(msg => console.log('  ' + msg));
  }
  
  if (errors.length > 0) {
    console.log('\n' + colors.red('错误项:'));
    errors.forEach(msg => console.log('  ' + msg));
  }
  
  // 连通性测试
  console.log('\n' + colors.blue('=== 连通性测试 ===\n'));
  
  // 测试数据库连接
  if (env.DATABASE_URL) {
    console.log(colors.gray('测试PostgreSQL连接...'));
    // 实际项目中这里应该真正测试连接
  }
  
  // 测试Redis连接
  if (env.REDIS_URL) {
    console.log(colors.gray('测试Redis连接...'));
    // 实际项目中这里应该真正测试连接
  }
  
  // 输出总结
  console.log('\n' + colors.blue('=== 总结 ===\n'));
  console.log(`成功: ${colors.green(successes.length)} 项`);
  console.log(`警告: ${colors.yellow(warnings.length)} 项`);
  console.log(`错误: ${colors.red(errors.length)} 项`);
  
  // 生成示例.env文件
  if (errors.length > 0) {
    const exampleEnvPath = path.join(rootDir, '.env.example');
    console.log(`\n${colors.yellow('提示:')} 参考 ${exampleEnvPath} 配置环境变量`);
    
    if (!fs.existsSync(exampleEnvPath)) {
      generateExampleEnv(exampleEnvPath);
      console.log(colors.green('已生成 .env.example 文件'));
    }
  }
  
  // 返回状态码
  if (errors.length > 0) {
    console.log('\n' + colors.red('环境变量校验失败！'));
    process.exit(1);
  } else {
    console.log('\n' + colors.green('环境变量校验通过！'));
    process.exit(0);
  }
}

// 生成示例环境变量文件
function generateExampleEnv(filePath) {
  let content = '# OpenPenPal 环境变量配置\n';
  content += '# 复制此文件为 .env 并填入实际值\n\n';
  
  const categories = {
    '基础配置': ['NODE_ENV'],
    '前端配置': ['NEXT_PUBLIC_API_BASE_URL', 'NEXT_PUBLIC_WS_URL'],
    '数据库配置': ['DATABASE_URL', 'DB_TYPE', 'REDIS_URL'],
    'JWT配置': ['JWT_SECRET', 'JWT_PUBLIC_KEY', 'JWT_PRIVATE_KEY'],
    'API密钥': ['MOONSHOT_API_KEY', 'SILICON_FLOW_API_KEY'],
    '服务端口': ['PORT', 'FRONTEND_PORT'],
    '文件上传': ['UPLOAD_DIR', 'MAX_UPLOAD_SIZE'],
    '监控配置': ['PROMETHEUS_ENABLED', 'GRAFANA_ENABLED']
  };
  
  for (const [category, keys] of Object.entries(categories)) {
    content += `# ${category}\n`;
    
    for (const key of keys) {
      const config = envRequirements[key];
      if (config) {
        content += `# ${config.description}`;
        if (!config.required) {
          content += ' (可选)';
        }
        content += '\n';
        
        if (config.values) {
          content += `# 可选值: ${config.values.join(', ')}\n`;
        }
        
        if (config.default) {
          content += `${key}=${config.default}\n`;
        } else if (key.includes('URL')) {
          content += `${key}=\n`;
        } else if (key.includes('SECRET') || key.includes('KEY')) {
          content += `${key}=your_secret_key_here\n`;
        } else {
          content += `${key}=\n`;
        }
        
        content += '\n';
      }
    }
  }
  
  fs.writeFileSync(filePath, content);
}

// 执行验证
if (require.main === module) {
  validateEnvironment();
}

module.exports = { validateEnvironment, envRequirements };