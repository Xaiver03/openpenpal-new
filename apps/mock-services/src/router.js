/**
 * 统一路由管理器
 * 自动加载和注册所有 API 路由
 */

import express from 'express';
import { requireAuth, requirePermissions, requireServiceAccess, optionalAuth } from './middleware/auth.js';
import { createLogger } from './utils/logger.js';

// 导入 API 处理函数
import * as authApi from './api/auth/login.js';
import * as letterApi from './api/write/letters.js';
import * as courierApi from './api/courier/tasks.js';

const logger = createLogger('router');

/**
 * 创建服务路由器
 */
export function createServiceRouter(serviceName) {
  const router = express.Router();
  
  logger.info(`创建 ${serviceName} 服务路由`);
  
  switch (serviceName) {
    case 'auth':
      setupAuthRoutes(router);
      break;
    case 'write-service':
      setupWriteServiceRoutes(router);
      break;
    case 'courier-service':
      setupCourierServiceRoutes(router);
      break;
    case 'admin-service':
      setupAdminServiceRoutes(router);
      break;
    case 'main-backend':
      setupMainBackendRoutes(router);
      break;
    case 'ocr-service':
      setupOcrServiceRoutes(router);
      break;
    default:
      logger.warn(`未知服务: ${serviceName}`);
  }
  
  return router;
}

/**
 * 设置认证服务路由
 */
function setupAuthRoutes(router) {
  // 公开路由 - 不需要认证
  router.post('/login', authApi.login);
  router.post('/register', authApi.register);
  
  // 需要认证的路由
  router.post('/refresh', requireAuth, authApi.refreshToken);
  router.post('/logout', requireAuth, authApi.logout);
  router.get('/me', requireAuth, authApi.getCurrentUser);
  
  logger.info('认证服务路由设置完成');
}

/**
 * 设置写信服务路由
 */
function setupWriteServiceRoutes(router) {
  // 公开路由 - 不需要认证 (放在认证中间件之前)
  router.get('/letters/public', letterApi.getPublicLetters);
  router.get('/api/v1/letters/public', letterApi.getPublicLetters);
  
  // 需要认证的路由
  router.use(requireAuth);
  router.use(requireServiceAccess('write-service'));
  
  // 信件相关路由
  router.get('/letters', letterApi.getLetters);
  router.post('/letters', requirePermissions(['WRITE_CREATE']), letterApi.createLetter);
  router.get('/letters/stats', letterApi.getLetterStats);
  router.get('/letters/:id', letterApi.getLetterById);
  router.put('/letters/:id/status', requirePermissions(['WRITE_UPDATE', 'ADMIN_WRITE']), letterApi.updateLetterStatus);
  router.delete('/letters/:id', requirePermissions(['WRITE_DELETE', 'ADMIN_WRITE']), letterApi.deleteLetter);
  
  logger.info('写信服务路由设置完成');
}

/**
 * 设置信使服务路由
 */
function setupCourierServiceRoutes(router) {
  // 所有信使服务的路由都需要认证
  router.use(requireAuth);
  router.use(requireServiceAccess('courier-service'));
  
  // 任务相关路由
  router.get('/tasks', courierApi.getAvailableTasks);
  router.get('/tasks/my', requirePermissions(['COURIER_READ']), courierApi.getMyTasks);
  router.post('/tasks/:id/accept', requirePermissions(['TASK_ACCEPT']), courierApi.acceptTask);
  router.put('/tasks/:id/status', requirePermissions(['TASK_COMPLETE']), courierApi.updateTaskStatus);
  
  // 信使相关路由
  router.get('/courier/stats', requirePermissions(['COURIER_READ']), courierApi.getCourierStats);
  router.post('/courier/apply', courierApi.applyCourier);
  
  // 信使管理路由 - 创建下级信使
  router.post('/courier/create', requirePermissions(['COURIER_MANAGE']), courierApi.createCourier);
  router.get('/courier/creatable-levels', requirePermissions(['COURIER_READ']), courierApi.getCreatableLevels);
  
  // 信使层级查看路由
  router.get('/courier/subordinates/list', requirePermissions(['COURIER_READ']), courierApi.getSubordinateCouriersList);
  router.get('/courier/:id/details', requirePermissions(['COURIER_READ']), courierApi.getCourierDetails);
  
  // 信使修改和审批路由
  router.put('/courier/:id/modify', requirePermissions(['COURIER_MANAGE']), courierApi.modifyCourier);
  router.get('/courier/approvals/pending', requirePermissions(['COURIER_READ']), courierApi.getPendingApprovals);
  router.post('/courier/approvals/:id/review', requirePermissions(['COURIER_MANAGE']), courierApi.approveModification);
  router.get('/courier/modifications/my-requests', requirePermissions(['COURIER_READ']), courierApi.getMyModificationRequests);
  
  logger.info('信使服务路由设置完成');
}

/**
 * 设置管理服务路由
 */
function setupAdminServiceRoutes(router) {
  // 管理服务需要管理员权限
  router.use(requireAuth);
  router.use(requireServiceAccess('admin-service'));
  
  // 用户管理
  router.get('/users', requirePermissions(['USER_MANAGE']), async (req, res) => {
    const { page = 0, size = 20, search, role, schoolCode, status } = req.query;
    
    // Mock 用户列表数据
    const mockUsers = [
      {
        id: 'user_001',
        username: 'alice',
        email: 'alice@pku.edu.cn',
        role: 'student',
        schoolCode: 'PKU',
        status: 'ACTIVE',
        profile: { fullName: '爱丽丝', grade: '大二' },
        createdAt: '2024-01-10T00:00:00Z',
        lastLogin: '2024-01-20T10:30:00Z'
      },
      {
        id: 'user_002',
        username: 'bob',
        email: 'bob@tsinghua.edu.cn',
        role: 'student',
        schoolCode: 'THU',
        status: 'ACTIVE',
        profile: { fullName: '鲍勃', grade: '大三' },
        createdAt: '2024-01-12T00:00:00Z',
        lastLogin: '2024-01-19T15:45:00Z'
      }
    ];
    
    return res.paginated(mockUsers, {
      page: parseInt(page),
      limit: parseInt(size),
      total: mockUsers.length
    });
  });
  
  // 信件管理
  router.get('/letters', requirePermissions(['ADMIN_READ']), async (req, res) => {
    return res.paginated([], { page: 0, limit: 20, total: 0 });
  });
  
  // 系统配置
  router.get('/system/config', requirePermissions(['SYSTEM_CONFIG']), async (req, res) => {
    const mockConfig = {
      maxLetterLength: 2000,
      deliveryTimeout: 72,
      autoMatchEnabled: true,
      maintenanceMode: false
    };
    return res.success(mockConfig);
  });
  
  // 博物馆管理
  router.get('/museum/exhibitions', requirePermissions(['MUSEUM_MANAGE']), async (req, res) => {
    const mockExhibitions = [
      {
        id: 'exhibition_001',
        title: '冬日温暖信件展',
        description: '收录冬季主题的温暖信件',
        status: 'active',
        letterCount: 15,
        createdAt: '2024-01-15T00:00:00Z'
      }
    ];
    return res.paginated(mockExhibitions);
  });
  
  router.get('/museum/moderation/tasks', requirePermissions(['CONTENT_MODERATE']), async (req, res) => {
    return res.paginated([], { page: 0, limit: 20, total: 0 });
  });
  
  logger.info('管理服务路由设置完成');
}

/**
 * 设置主后端服务路由
 */
function setupMainBackendRoutes(router) {
  router.use(requireAuth);
  
  // 用户资料相关
  router.get('/users/profile', async (req, res) => {
    const { password: _, ...userProfile } = req.user;
    return res.success(userProfile);
  });
  
  router.put('/users/profile', requirePermissions(['PROFILE_UPDATE']), async (req, res) => {
    const { fullName, bio, avatar } = req.body;
    
    // Mock 更新用户资料
    const updatedProfile = {
      ...req.user.profile,
      fullName: fullName || req.user.profile.fullName,
      bio: bio || req.user.profile.bio,
      avatar: avatar || req.user.profile.avatar,
      updatedAt: new Date().toISOString()
    };
    
    return res.success(updatedProfile, '用户资料更新成功');
  });
  
  // 健康检查
  router.get('/health', optionalAuth, async (req, res) => {
    return res.success({
      status: 'healthy',
      timestamp: new Date().toISOString(),
      service: 'main-backend',
      version: '1.0.0'
    });
  });
  
  logger.info('主后端服务路由设置完成');
}

/**
 * 设置 OCR 服务路由
 */
function setupOcrServiceRoutes(router) {
  router.use(requireAuth);
  
  // 健康检查
  router.get('/health', async (req, res) => {
    return res.success({
      status: 'healthy',
      timestamp: new Date().toISOString(),
      service: 'ocr-service',
      version: '1.0.0',
      models: ['general', 'handwriting', 'printed']
    });
  });
  
  // OCR 模型列表
  router.get('/models', async (req, res) => {
    const models = [
      {
        id: 'general',
        name: '通用文字识别',
        description: '适用于各种类型的文字识别',
        accuracy: 0.95,
        supportedLanguages: ['zh-CN', 'en']
      },
      {
        id: 'handwriting',
        name: '手写文字识别',
        description: '专门用于手写文字的识别',
        accuracy: 0.88,
        supportedLanguages: ['zh-CN']
      }
    ];
    
    return res.success(models);
  });
  
  // OCR 处理（需要权限）
  router.post('/process', requirePermissions(['OCR_PROCESS']), async (req, res) => {
    const { imageUrl, modelId = 'general' } = req.body;
    
    if (!imageUrl) {
      return res.validationError([
        { field: 'imageUrl', message: '图片URL不能为空' }
      ]);
    }
    
    // Mock OCR 处理结果
    const result = {
      id: `ocr_${Date.now()}`,
      status: 'completed',
      text: '这是识别出的文字内容示例。\n这是第二行文字。',
      confidence: 0.92,
      modelUsed: modelId,
      processingTime: 1.5,
      boundingBoxes: [
        {
          text: '这是识别出的文字内容示例。',
          coordinates: { x: 10, y: 20, width: 300, height: 25 },
          confidence: 0.95
        }
      ],
      createdAt: new Date().toISOString()
    };
    
    return res.success(result, 'OCR 处理完成');
  });
  
  logger.info('OCR 服务路由设置完成');
}

/**
 * 创建通用健康检查路由
 */
export function createHealthRouter() {
  const router = express.Router();
  
  router.get('/', (req, res) => {
    res.success({
      status: 'healthy',
      timestamp: new Date().toISOString(),
      service: 'mock-services',
      version: '1.0.0',
      uptime: process.uptime()
    });
  });
  
  return router;
}