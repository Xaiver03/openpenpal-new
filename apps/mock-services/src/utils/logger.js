/**
 * 统一日志工具
 * 提供彩色输出、不同级别的日志记录
 */

// 颜色定义
const colors = {
  reset: '\x1b[0m',
  bright: '\x1b[1m',
  dim: '\x1b[2m',
  red: '\x1b[31m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  magenta: '\x1b[35m',
  cyan: '\x1b[36m',
  white: '\x1b[37m',
  gray: '\x1b[90m'
};

// 日志级别
const LOG_LEVELS = {
  debug: 0,
  info: 1,
  warn: 2,
  error: 3
};

// 当前日志级别
let currentLogLevel = LOG_LEVELS.info;

/**
 * 设置日志级别
 */
export function setLogLevel(level) {
  if (LOG_LEVELS.hasOwnProperty(level)) {
    currentLogLevel = LOG_LEVELS[level];
  }
}

/**
 * 格式化时间戳
 */
function formatTimestamp() {
  const now = new Date();
  return now.toISOString().replace('T', ' ').slice(0, 19);
}

/**
 * 格式化日志消息
 */
function formatMessage(level, module, message, data = null) {
  const timestamp = formatTimestamp();
  const moduleStr = module ? `[${module}]` : '';
  const dataStr = data ? `\n${JSON.stringify(data, null, 2)}` : '';
  
  return `${timestamp} ${moduleStr} ${message}${dataStr}`;
}

/**
 * 输出彩色日志
 */
function logWithColor(level, color, module, message, data) {
  if (LOG_LEVELS[level] < currentLogLevel) {
    return;
  }
  
  const formattedMessage = formatMessage(level, module, message, data);
  const coloredMessage = `${color}${formattedMessage}${colors.reset}`;
  
  if (level === 'error') {
    console.error(coloredMessage);
  } else if (level === 'warn') {
    console.warn(coloredMessage);
  } else {
    console.log(coloredMessage);
  }
}

/**
 * 创建模块化日志记录器
 */
export function createLogger(module = '') {
  return {
    debug: (message, data) => logWithColor('debug', colors.gray, module, `[DEBUG] ${message}`, data),
    info: (message, data) => logWithColor('info', colors.blue, module, `[INFO] ${message}`, data),
    warn: (message, data) => logWithColor('warn', colors.yellow, module, `[WARN] ${message}`, data),
    error: (message, data) => logWithColor('error', colors.red, module, `[ERROR] ${message}`, data),
    success: (message, data) => logWithColor('info', colors.green, module, `[SUCCESS] ${message}`, data),
    
    // 通用日志方法
    log: (level, message, data) => {
      const levelColors = {
        debug: colors.gray,
        info: colors.blue,
        warn: colors.yellow,
        error: colors.red
      };
      logWithColor(level, levelColors[level] || colors.white, module, `[${level.toUpperCase()}] ${message}`, data);
    }
  };
}

/**
 * 全局日志记录器
 */
export const logger = createLogger();

/**
 * 请求日志格式化器
 */
export function formatRequestLog(req, res, duration) {
  const method = req.method;
  const url = req.url;
  const status = res.statusCode;
  const user = req.user ? req.user.username : 'anonymous';
  const ip = req.ip || req.connection.remoteAddress;
  
  // 根据状态码选择颜色
  let statusColor = colors.green;
  if (status >= 400 && status < 500) {
    statusColor = colors.yellow;
  } else if (status >= 500) {
    statusColor = colors.red;
  }
  
  return {
    message: `${method} ${url} - ${status} (${duration}ms) - ${user}@${ip}`,
    color: statusColor
  };
}

/**
 * 错误日志格式化器
 */
export function formatErrorLog(error, req = null) {
  const stack = error.stack || '';
  const message = error.message || 'Unknown error';
  
  let context = '';
  if (req) {
    context = `[${req.method} ${req.url}]`;
  }
  
  return {
    message: `${context} ${message}`,
    stack: stack
  };
}

/**
 * 启动信息日志
 */
export function logStartup(serviceName, port, config = {}) {
  const logger = createLogger('startup');
  
  logger.info(`🚀 ${serviceName} Mock Service 启动中...`);
  logger.info(`📍 端口: ${port}`);
  
  if (config.cors) {
    logger.info(`🌍 CORS 已启用: ${config.cors.origin.join(', ')}`);
  }
  
  if (config.auth) {
    logger.info(`🔐 认证: ${config.auth ? '已启用' : '已禁用'}`);
  }
  
  if (config.delay) {
    logger.info(`⏱️  延迟模拟: ${config.delay ? '已启用' : '已禁用'}`);
  }
  
  if (config.errors) {
    logger.info(`💥 错误模拟: ${config.errors ? '已启用' : '已禁用'}`);
  }
  
  logger.success(`✅ ${serviceName} 服务已启动: http://localhost:${port}`);
}

/**
 * 关闭信息日志
 */
export function logShutdown(serviceName) {
  const logger = createLogger('shutdown');
  logger.info(`🛑 ${serviceName} Mock Service 正在关闭...`);
  logger.success(`✅ ${serviceName} 服务已安全关闭`);
}