// 统一配置管理器
const fs = require('fs');
const path = require('path');
const dotenv = require('dotenv');
const Ajv = require('ajv');
const schema = require('./config.schema.json');

class ConfigManager {
  constructor() {
    this.config = {};
    this.env = process.env.NODE_ENV || 'development';
    this.loadEnvironmentVariables();
    this.loadConfigFiles();
    this.validateConfig();
  }

  loadEnvironmentVariables() {
    // 加载基础 .env 文件
    const envPath = path.resolve(process.cwd(), '.env');
    if (fs.existsSync(envPath)) {
      dotenv.config({ path: envPath });
    }

    // 加载环境特定的 .env 文件
    const envSpecificPath = path.resolve(process.cwd(), `.env.${this.env}`);
    if (fs.existsSync(envSpecificPath)) {
      dotenv.config({ path: envSpecificPath });
    }

    // 加载本地覆盖文件
    const localPath = path.resolve(process.cwd(), '.env.local');
    if (fs.existsSync(localPath)) {
      dotenv.config({ path: localPath });
    }
  }

  loadConfigFiles() {
    // 加载默认配置
    const defaultConfig = this.loadJsonFile('config.default.json');
    
    // 加载环境特定配置
    const envConfig = this.loadJsonFile(`config.${this.env}.json`);
    
    // 合并配置
    this.config = this.deepMerge(defaultConfig, envConfig);
    
    // 应用环境变量覆盖
    this.applyEnvironmentOverrides();
  }

  loadJsonFile(filename) {
    const filepath = path.join(__dirname, filename);
    if (fs.existsSync(filepath)) {
      return JSON.parse(fs.readFileSync(filepath, 'utf8'));
    }
    return {};
  }

  deepMerge(target, source) {
    const output = Object.assign({}, target);
    if (this.isObject(target) && this.isObject(source)) {
      Object.keys(source).forEach(key => {
        if (this.isObject(source[key])) {
          if (!(key in target)) {
            Object.assign(output, { [key]: source[key] });
          } else {
            output[key] = this.deepMerge(target[key], source[key]);
          }
        } else {
          Object.assign(output, { [key]: source[key] });
        }
      });
    }
    return output;
  }

  isObject(item) {
    return item && typeof item === 'object' && !Array.isArray(item);
  }

  applyEnvironmentOverrides() {
    // 映射环境变量到配置
    const envMappings = {
      'APP_NAME': 'app.name',
      'APP_URL': 'app.url',
      'APP_PORT': 'app.port',
      'NODE_ENV': 'app.env',
      'DATABASE_URL': 'database.url',
      'REDIS_URL': 'redis.url',
      'REDIS_PASSWORD': 'redis.password',
      'JWT_SECRET': 'auth.jwt.secret',
      'JWT_EXPIRES_IN': 'auth.jwt.expiresIn',
      'MAIL_HOST': 'mail.host',
      'MAIL_PORT': 'mail.port',
      'MAIL_USERNAME': 'mail.auth.user',
      'MAIL_PASSWORD': 'mail.auth.pass',
      'MAIL_FROM_ADDRESS': 'mail.from.address',
      'MAIL_FROM_NAME': 'mail.from.name',
      'STORAGE_TYPE': 'storage.type',
      'SENTRY_DSN': 'monitoring.sentry.dsn',
    };

    Object.entries(envMappings).forEach(([envKey, configPath]) => {
      if (process.env[envKey]) {
        this.setConfigValue(configPath, process.env[envKey]);
      }
    });

    // 处理端口配置
    const portMappings = {
      'PORT_FRONTEND': 'services.frontend.port',
      'PORT_API_GATEWAY': 'services.apiGateway.port',
      'PORT_WRITE_SERVICE': 'services.writeService.port',
      'PORT_COURIER_SERVICE': 'services.courierService.port',
      'PORT_ADMIN_SERVICE': 'services.adminService.port',
      'PORT_OCR_SERVICE': 'services.ocrService.port',
    };

    Object.entries(portMappings).forEach(([envKey, configPath]) => {
      if (process.env[envKey]) {
        this.setConfigValue(configPath, parseInt(process.env[envKey], 10));
      }
    });
  }

  setConfigValue(path, value) {
    const keys = path.split('.');
    let current = this.config;
    
    for (let i = 0; i < keys.length - 1; i++) {
      if (!current[keys[i]]) {
        current[keys[i]] = {};
      }
      current = current[keys[i]];
    }
    
    current[keys[keys.length - 1]] = value;
  }

  validateConfig() {
    const ajv = new Ajv({ useDefaults: true });
    const validate = ajv.compile(schema);
    
    if (!validate(this.config)) {
      console.error('配置验证失败:', validate.errors);
      throw new Error('Invalid configuration');
    }
  }

  get(path, defaultValue = undefined) {
    const keys = path.split('.');
    let current = this.config;
    
    for (const key of keys) {
      if (current[key] === undefined) {
        return defaultValue;
      }
      current = current[key];
    }
    
    return current;
  }

  getAll() {
    return this.config;
  }

  isProduction() {
    return this.env === 'production';
  }

  isDevelopment() {
    return this.env === 'development';
  }

  isTest() {
    return this.env === 'test';
  }
}

// 单例模式
let instance = null;

module.exports = {
  config: () => {
    if (!instance) {
      instance = new ConfigManager();
    }
    return instance;
  },
  
  // 便捷方法
  get: (path, defaultValue) => {
    if (!instance) {
      instance = new ConfigManager();
    }
    return instance.get(path, defaultValue);
  }
};