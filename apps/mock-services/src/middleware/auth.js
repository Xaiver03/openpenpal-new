/**
 * 认证和权限中间件
 * 统一处理 JWT token 验证和权限检查
 */

import jwt from 'jsonwebtoken';
import { DEFAULT_CONFIG } from '../config/services.js';
import { findUserById, hasPermission, canAccessService } from '../config/users.js';
import { createLogger } from '../utils/logger.js';

const logger = createLogger('auth');

/**
 * 生成 JWT Token
 */
export function generateToken(user) {
  const payload = {
    id: user.id,
    username: user.username,
    role: user.role,
    schoolCode: user.schoolCode,
    permissions: user.permissions
  };
  
  return jwt.sign(payload, DEFAULT_CONFIG.jwt.secret, {
    expiresIn: DEFAULT_CONFIG.jwt.expiresIn,
    issuer: DEFAULT_CONFIG.jwt.issuer,
    audience: DEFAULT_CONFIG.jwt.audience
  });
}

/**
 * 验证 JWT Token
 */
export function verifyToken(token) {
  try {
    const decoded = jwt.verify(token, DEFAULT_CONFIG.jwt.secret, {
      issuer: DEFAULT_CONFIG.jwt.issuer,
      audience: DEFAULT_CONFIG.jwt.audience
    });
    return decoded;
  } catch (error) {
    logger.warn('Token 验证失败:', error.message);
    return null;
  }
}

/**
 * 认证中间件 - 验证用户是否已登录
 */
export function requireAuth(req, res, next) {
  const authHeader = req.headers.authorization;
  
  if (!authHeader || !authHeader.startsWith('Bearer ')) {
    return res.status(401).json({
      code: 401,
      msg: '缺少认证 token',
      error: {
        type: 'AUTHENTICATION_REQUIRED',
        details: '请在请求头中提供有效的 Authorization Bearer token'
      },
      timestamp: new Date().toISOString()
    });
  }
  
  const token = authHeader.substring(7); // 移除 'Bearer ' 前缀
  const decoded = verifyToken(token);
  
  if (!decoded) {
    return res.status(401).json({
      code: 401,
      msg: 'Token 无效或已过期',
      error: {
        type: 'INVALID_TOKEN',
        details: '请重新登录获取有效 token'
      },
      timestamp: new Date().toISOString()
    });
  }
  
  // 从数据库中获取最新用户信息（模拟）
  const user = findUserById(decoded.id);
  if (!user) {
    return res.status(401).json({
      code: 401,
      msg: '用户不存在',
      error: {
        type: 'USER_NOT_FOUND',
        details: '用户可能已被删除'
      },
      timestamp: new Date().toISOString()
    });
  }
  
  // 将用户信息附加到请求对象
  req.user = user;
  req.token = decoded;
  
  logger.debug(`用户 ${user.username} 通过认证`);
  next();
}

/**
 * 权限检查中间件工厂函数
 * @param {string|string[]} requiredPermissions - 必需的权限
 * @returns {Function} Express 中间件函数
 */
export function requirePermissions(requiredPermissions) {
  const permissions = Array.isArray(requiredPermissions) ? requiredPermissions : [requiredPermissions];
  
  return (req, res, next) => {
    if (!req.user) {
      return res.status(401).json({
        code: 401,
        msg: '用户未认证',
        error: {
          type: 'USER_NOT_AUTHENTICATED',
          details: '请先通过认证中间件'
        },
        timestamp: new Date().toISOString()
      });
    }
    
    // 检查用户是否具有所需的任一权限
    const hasRequiredPermission = permissions.some(permission => 
      hasPermission(req.user, permission)
    );
    
    if (!hasRequiredPermission) {
      logger.warn(`用户 ${req.user.username} 缺少权限:`, permissions);
      return res.status(403).json({
        code: 403,
        msg: '权限不足',
        error: {
          type: 'INSUFFICIENT_PERMISSIONS',
          details: `需要以下权限之一: ${permissions.join(', ')}`,
          required: permissions,
          current: req.user.permissions
        },
        timestamp: new Date().toISOString()
      });
    }
    
    logger.debug(`用户 ${req.user.username} 权限检查通过:`, permissions);
    next();
  };
}

/**
 * 服务访问权限中间件工厂函数
 * @param {string} serviceName - 服务名称
 * @returns {Function} Express 中间件函数
 */
export function requireServiceAccess(serviceName) {
  return (req, res, next) => {
    if (!req.user) {
      return res.status(401).json({
        code: 401,
        msg: '用户未认证',
        timestamp: new Date().toISOString()
      });
    }
    
    if (!canAccessService(req.user, serviceName)) {
      logger.warn(`用户 ${req.user.username} 无法访问服务:`, serviceName);
      return res.status(403).json({
        code: 403,
        msg: `无权访问 ${serviceName} 服务`,
        error: {
          type: 'SERVICE_ACCESS_DENIED',
          details: `用户角色 ${req.user.role} 无权访问此服务`,
          service: serviceName
        },
        timestamp: new Date().toISOString()
      });
    }
    
    logger.debug(`用户 ${req.user.username} 获得服务访问权限:`, serviceName);
    next();
  };
}

/**
 * 角色检查中间件工厂函数
 * @param {string|string[]} requiredRoles - 必需的角色
 * @returns {Function} Express 中间件函数
 */
export function requireRoles(requiredRoles) {
  const roles = Array.isArray(requiredRoles) ? requiredRoles : [requiredRoles];
  
  return (req, res, next) => {
    if (!req.user) {
      return res.status(401).json({
        code: 401,
        msg: '用户未认证',
        timestamp: new Date().toISOString()
      });
    }
    
    if (!roles.includes(req.user.role)) {
      logger.warn(`用户 ${req.user.username} 角色不匹配:`, {
        required: roles,
        current: req.user.role
      });
      return res.status(403).json({
        code: 403,
        msg: '角色权限不足',
        error: {
          type: 'ROLE_ACCESS_DENIED',
          details: `需要以下角色之一: ${roles.join(', ')}`,
          required: roles,
          current: req.user.role
        },
        timestamp: new Date().toISOString()
      });
    }
    
    logger.debug(`用户 ${req.user.username} 角色检查通过:`, req.user.role);
    next();
  };
}

/**
 * 可选认证中间件 - 如果有 token 则验证，没有则跳过
 */
export function optionalAuth(req, res, next) {
  const authHeader = req.headers.authorization;
  
  if (authHeader && authHeader.startsWith('Bearer ')) {
    const token = authHeader.substring(7);
    const decoded = verifyToken(token);
    
    if (decoded) {
      const user = findUserById(decoded.id);
      if (user) {
        req.user = user;
        req.token = decoded;
        logger.debug(`可选认证成功: ${user.username}`);
      }
    }
  }
  
  next();
}