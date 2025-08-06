/**
 * 响应处理中间件
 * 统一处理响应格式、延迟模拟、错误模拟等
 */

import { DEFAULT_CONFIG } from '../config/services.js';
import { createLogger } from '../utils/logger.js';

const logger = createLogger('response');

/**
 * 延迟模拟中间件
 */
export function simulateDelay(serviceName) {
  return async (req, res, next) => {
    // 检查是否启用了延迟模拟
    let delay = 0;
    
    // 优先使用查询参数中的延迟设置
    if (req.query.delay) {
      delay = parseInt(req.query.delay, 10);
      if (isNaN(delay) || delay < 0) {
        delay = 0;
      }
    } else if (DEFAULT_CONFIG.globalDelay.enabled) {
      // 使用全局延迟配置
      const { min, max } = DEFAULT_CONFIG.globalDelay;
      delay = Math.floor(Math.random() * (max - min + 1)) + min;
    }
    
    if (delay > 0) {
      logger.debug(`${serviceName} 模拟延迟: ${delay}ms`);
      await new Promise(resolve => setTimeout(resolve, delay));
    }
    
    next();
  };
}

/**
 * 错误模拟中间件
 */
export function simulateErrors() {
  return (req, res, next) => {
    if (!DEFAULT_CONFIG.errorSimulation.enabled) {
      return next();
    }
    
    const shouldSimulateError = Math.random() < DEFAULT_CONFIG.errorSimulation.probability;
    
    if (shouldSimulateError) {
      const errorTypes = DEFAULT_CONFIG.errorSimulation.types;
      const randomErrorType = errorTypes[Math.floor(Math.random() * errorTypes.length)];
      
      logger.warn(`模拟 ${randomErrorType} 错误`);
      
      switch (randomErrorType) {
        case 'network':
          return res.status(503).json({
            code: 503,
            msg: '网络连接错误',
            error: {
              type: 'NETWORK_ERROR',
              details: '模拟网络不稳定导致的连接失败'
            },
            timestamp: new Date().toISOString()
          });
          
        case 'server':
          return res.status(500).json({
            code: 500,
            msg: '服务器内部错误',
            error: {
              type: 'INTERNAL_SERVER_ERROR',
              details: '模拟服务器异常'
            },
            timestamp: new Date().toISOString()
          });
          
        case 'timeout':
          // 模拟超时，实际上是延迟很长时间后返回错误
          setTimeout(() => {
            if (!res.headersSent) {
              res.status(408).json({
                code: 408,
                msg: '请求超时',
                error: {
                  type: 'REQUEST_TIMEOUT',
                  details: '模拟请求处理超时'
                },
                timestamp: new Date().toISOString()
              });
            }
          }, 5000);
          return;
          
        default:
          break;
      }
    }
    
    next();
  };
}

/**
 * 统一响应格式中间件
 */
export function formatResponse() {
  return (req, res, next) => {
    // 扩展 res 对象，添加统一的响应方法
    
    // 成功响应
    res.success = function(data = null, message = '操作成功') {
      const response = {
        code: 0,
        msg: message,
        data,
        timestamp: new Date().toISOString()
      };
      
      // 添加分页信息（如果存在）
      if (req.pagination) {
        response.pagination = req.pagination;
      }
      
      logger.debug('成功响应:', { url: req.url, method: req.method });
      return this.json(response);
    };
    
    // 错误响应
    res.error = function(code = 500, message = '操作失败', details = null) {
      const response = {
        code,
        msg: message,
        error: {
          type: 'OPERATION_FAILED',
          details: details || message,
          path: req.path,
          method: req.method
        },
        timestamp: new Date().toISOString()
      };
      
      logger.error('错误响应:', { 
        url: req.url, 
        method: req.method, 
        code, 
        message 
      });
      
      return this.status(code >= 100 && code < 600 ? code : 500).json(response);
    };
    
    // 分页响应
    res.paginated = function(items = [], pagination = {}) {
      const defaultPagination = {
        page: 0,
        limit: 20,
        total: items.length,
        pages: Math.ceil(items.length / (pagination.limit || 20)),
        hasNext: false,
        hasPrev: false
      };
      
      const finalPagination = { ...defaultPagination, ...pagination };
      finalPagination.hasNext = finalPagination.page < finalPagination.pages - 1;
      finalPagination.hasPrev = finalPagination.page > 0;
      
      return this.success({
        items,
        pagination: finalPagination
      });
    };
    
    // 验证错误响应
    res.validationError = function(errors = []) {
      return this.status(400).json({
        code: 400,
        msg: '数据验证失败',
        error: {
          type: 'VALIDATION_ERROR',
          details: '请检查输入数据格式',
          fields: errors
        },
        timestamp: new Date().toISOString()
      });
    };
    
    next();
  };
}

/**
 * 请求日志中间件
 */
export function requestLogger() {
  return (req, res, next) => {
    const start = Date.now();
    
    // 记录请求开始
    logger.info(`${req.method} ${req.url}`, {
      ip: req.ip,
      userAgent: req.get('User-Agent'),
      user: req.user ? req.user.username : 'anonymous'
    });
    
    // 监听响应结束事件
    res.on('finish', () => {
      const duration = Date.now() - start;
      const level = res.statusCode >= 400 ? 'warn' : 'info';
      
      logger.log(level, `${req.method} ${req.url} - ${res.statusCode} (${duration}ms)`, {
        statusCode: res.statusCode,
        duration,
        user: req.user ? req.user.username : 'anonymous'
      });
    });
    
    next();
  };
}

/**
 * CORS 中间件（如果需要自定义 CORS 逻辑）
 */
export function customCors() {
  return (req, res, next) => {
    const origin = req.headers.origin;
    const allowedOrigins = DEFAULT_CONFIG.cors.origin;
    
    if (allowedOrigins.includes(origin) || allowedOrigins.includes('*')) {
      res.header('Access-Control-Allow-Origin', origin);
    }
    
    res.header('Access-Control-Allow-Credentials', 'true');
    res.header('Access-Control-Allow-Methods', DEFAULT_CONFIG.cors.methods.join(', '));
    res.header('Access-Control-Allow-Headers', DEFAULT_CONFIG.cors.allowedHeaders.join(', '));
    
    // 处理预检请求
    if (req.method === 'OPTIONS') {
      return res.status(200).end();
    }
    
    next();
  };
}