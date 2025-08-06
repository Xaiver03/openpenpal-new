/**
 * 认证相关 API
 * 处理登录、注册、token 刷新等认证操作
 */

import bcrypt from 'bcryptjs';
import { findUserByUsername } from '../../config/users.js';
import { generateToken } from '../../middleware/auth.js';
import { createLogger } from '../../utils/logger.js';

const logger = createLogger('auth-api');

/**
 * 用户登录
 */
export async function login(req, res) {
  try {
    const { username, password } = req.body;
    
    // 验证输入
    if (!username || !password) {
      return res.validationError([
        { field: 'username', message: '用户名不能为空' },
        { field: 'password', message: '密码不能为空' }
      ]);
    }
    
    // 查找用户
    const user = findUserByUsername(username);
    if (!user) {
      logger.warn(`登录失败 - 用户不存在: ${username}`);
      return res.error(401, '用户名或密码错误');
    }
    
    // 验证密码（在真实环境中应该使用 bcrypt）
    // 这里为了简化 mock，直接比较明文密码
    if (user.password !== password) {
      logger.warn(`登录失败 - 密码错误: ${username}`);
      return res.error(401, '用户名或密码错误');
    }
    
    // 生成 token
    const token = generateToken(user);
    
    // 返回用户信息（不包含密码）
    const { password: _, ...userInfo } = user;
    
    logger.info(`用户登录成功: ${username}`);
    
    return res.success({
      token,
      user: userInfo,
      expiresIn: '24h'
    }, '登录成功');
    
  } catch (error) {
    logger.error('登录处理异常:', error);
    return res.error(500, '登录失败，请稍后重试');
  }
}

/**
 * 用户注册
 */
export async function register(req, res) {
  try {
    const { username, email, password, schoolCode } = req.body;
    
    // 验证输入
    const errors = [];
    if (!username) errors.push({ field: 'username', message: '用户名不能为空' });
    if (!email) errors.push({ field: 'email', message: '邮箱不能为空' });
    if (!password) errors.push({ field: 'password', message: '密码不能为空' });
    if (!schoolCode) errors.push({ field: 'schoolCode', message: '学校代码不能为空' });
    
    if (errors.length > 0) {
      return res.validationError(errors);
    }
    
    // 检查用户是否已存在
    const existingUser = findUserByUsername(username);
    if (existingUser) {
      return res.error(409, '用户名已存在');
    }
    
    // 在真实环境中，这里应该：
    // 1. 加密密码
    // 2. 保存到数据库
    // 3. 发送验证邮件等
    
    // Mock 环境下直接返回成功
    logger.info(`新用户注册: ${username} (${email})`);
    
    return res.success({
      message: '注册成功，请等待管理员审核'
    }, '注册申请已提交');
    
  } catch (error) {
    logger.error('注册处理异常:', error);
    return res.error(500, '注册失败，请稍后重试');
  }
}

/**
 * 刷新 token
 */
export async function refreshToken(req, res) {
  try {
    // 从当前用户信息生成新 token
    if (!req.user) {
      return res.error(401, '用户未认证');
    }
    
    const newToken = generateToken(req.user);
    
    logger.info(`Token 刷新成功: ${req.user.username}`);
    
    return res.success({
      token: newToken,
      expiresIn: '24h'
    }, 'Token 刷新成功');
    
  } catch (error) {
    logger.error('Token 刷新异常:', error);
    return res.error(500, 'Token 刷新失败');
  }
}

/**
 * 用户登出
 */
export async function logout(req, res) {
  try {
    // 在真实环境中，这里应该：
    // 1. 将 token 加入黑名单
    // 2. 清理相关 session
    
    logger.info(`用户登出: ${req.user ? req.user.username : 'unknown'}`);
    
    return res.success(null, '登出成功');
    
  } catch (error) {
    logger.error('登出处理异常:', error);
    return res.error(500, '登出失败');
  }
}

/**
 * 获取当前用户信息
 */
export async function getCurrentUser(req, res) {
  try {
    if (!req.user) {
      return res.error(401, '用户未认证');
    }
    
    // 返回用户信息（不包含密码）
    const { password: _, ...userInfo } = req.user;
    
    return res.success(userInfo, '获取用户信息成功');
    
  } catch (error) {
    logger.error('获取用户信息异常:', error);
    return res.error(500, '获取用户信息失败');
  }
}