/**
 * 信使服务 API
 * 处理配送任务、信使管理等操作
 */

import { v4 as uuidv4 } from 'uuid';
import { createLogger } from '../../utils/logger.js';

const logger = createLogger('courier-api');

// Mock 配送任务数据
let deliveryTasks = [
  {
    id: 'task_001',
    letterId: 'letter_002',
    letterInfo: {
      title: '关于大学生活的思考',
      senderName: '鲍勃',
      senderSchool: 'THU'
    },
    pickupLocation: {
      name: '清华大学邮局',
      address: '清华大学主楼一层',
      coordinates: { lat: 40.0042, lng: 116.3264 },
      contactPerson: '张师傅',
      contactPhone: '010-62785001'
    },
    deliveryLocation: {
      name: '北京大学邮局',
      address: '北京大学理科楼群',
      coordinates: { lat: 39.9986, lng: 116.3060 },
      contactPerson: '李师傅',
      contactPhone: '010-62751234'
    },
    receiverHint: '北京大学计算机系的朋友',
    courierId: 'courier_001',
    courierInfo: {
      name: '快递员小王',
      phone: '13800138001',
      zone: '北京大学'
    },
    status: 'assigned', // available, assigned, picked_up, in_transit, delivered, completed
    priority: 'normal', // low, normal, high, urgent
    estimatedTime: 120, // 预计配送时间（分钟）
    reward: 15.00,
    deadline: new Date('2024-01-20T18:00:00Z').toISOString(),
    createdAt: new Date('2024-01-16T09:00:00Z').toISOString(),
    updatedAt: new Date('2024-01-16T10:30:00Z').toISOString(),
    timeline: [
      {
        status: 'available',
        timestamp: new Date('2024-01-16T09:00:00Z').toISOString(),
        note: '任务创建'
      },
      {
        status: 'assigned',
        timestamp: new Date('2024-01-16T10:30:00Z').toISOString(),
        note: '任务分配给快递员小王',
        operator: 'courier_001'
      }
    ]
  },
  {
    id: 'task_002',
    letterId: 'letter_001',
    letterInfo: {
      title: '给远方朋友的第一封信',
      senderName: '爱丽丝',
      senderSchool: 'PKU'
    },
    pickupLocation: {
      name: '北京大学邮局',
      address: '北京大学理科楼群',
      coordinates: { lat: 39.9986, lng: 116.3060 },
      contactPerson: '李师傅',
      contactPhone: '010-62751234'
    },
    deliveryLocation: {
      name: '清华大学邮局',
      address: '清华大学主楼一层',
      coordinates: { lat: 40.0042, lng: 116.3264 },
      contactPerson: '张师傅',
      contactPhone: '010-62785001'
    },
    receiverHint: '清华大学物理系的同学',
    courierId: null,
    courierInfo: null,
    status: 'available',
    priority: 'normal',
    estimatedTime: 90,
    reward: 12.00,
    deadline: new Date('2024-01-21T17:00:00Z').toISOString(),
    createdAt: new Date('2024-01-17T08:30:00Z').toISOString(),
    updatedAt: new Date('2024-01-17T08:30:00Z').toISOString(),
    timeline: [
      {
        status: 'available',
        timestamp: new Date('2024-01-17T08:30:00Z').toISOString(),
        note: '任务创建'
      }
    ]
  }
];

// Mock 信使数据
let couriers = [
  {
    id: 'courier_001',
    userId: 'courier_001',
    status: 'active', // active, inactive, busy, offline
    zone: '北京大学',
    rating: 4.8,
    completedTasks: 156,
    currentTasks: 1,
    maxTasks: 3,
    profile: {
      name: '快递员小王',
      phone: '13800138001',
      experience: '2年配送经验',
      avatar: '/avatars/courier1.png'
    },
    location: {
      coordinates: { lat: 39.9986, lng: 116.3060 },
      updateTime: new Date().toISOString()
    },
    workingHours: {
      start: '08:00',
      end: '20:00',
      timezone: 'Asia/Shanghai'
    }
  },
  {
    id: 'courier_002',
    userId: 'courier_002',
    status: 'active',
    zone: '清华大学',
    rating: 4.9,
    completedTasks: 203,
    currentTasks: 0,
    maxTasks: 3,
    profile: {
      name: '快递员小李',
      phone: '13800138002',
      experience: '3年配送经验',
      avatar: '/avatars/courier2.png'
    },
    location: {
      coordinates: { lat: 40.0042, lng: 116.3264 },
      updateTime: new Date().toISOString()
    },
    workingHours: {
      start: '09:00',
      end: '21:00',
      timezone: 'Asia/Shanghai'
    }
  }
];

/**
 * 获取可用任务列表
 */
export async function getAvailableTasks(req, res) {
  try {
    const { 
      page = 0, 
      limit = 20, 
      zone, 
      priority, 
      minReward,
      maxEstimatedTime 
    } = req.query;
    
    let availableTasks = deliveryTasks.filter(task => task.status === 'available');
    
    // 如果是信使用户，只显示其工作区域的任务
    if (req.user.role === 'courier') {
      const courier = couriers.find(c => c.userId === req.user.id);
      if (courier) {
        availableTasks = availableTasks.filter(task => 
          task.pickupLocation.name.includes(courier.zone) ||
          task.deliveryLocation.name.includes(courier.zone)
        );
      }
    }
    
    // 区域过滤
    if (zone) {
      availableTasks = availableTasks.filter(task => 
        task.pickupLocation.name.includes(zone) ||
        task.deliveryLocation.name.includes(zone)
      );
    }
    
    // 优先级过滤
    if (priority) {
      availableTasks = availableTasks.filter(task => task.priority === priority);
    }
    
    // 最低奖励过滤
    if (minReward) {
      const minRewardNum = parseFloat(minReward);
      availableTasks = availableTasks.filter(task => task.reward >= minRewardNum);
    }
    
    // 最大预计时间过滤
    if (maxEstimatedTime) {
      const maxTimeNum = parseInt(maxEstimatedTime, 10);
      availableTasks = availableTasks.filter(task => task.estimatedTime <= maxTimeNum);
    }
    
    // 按创建时间倒序排列
    availableTasks.sort((a, b) => new Date(b.createdAt) - new Date(a.createdAt));
    
    // 分页处理
    const pageNum = parseInt(page, 10);
    const limitNum = parseInt(limit, 10);
    const startIndex = pageNum * limitNum;
    const endIndex = startIndex + limitNum;
    
    const paginatedTasks = availableTasks.slice(startIndex, endIndex);
    
    const pagination = {
      page: pageNum,
      limit: limitNum,
      total: availableTasks.length,
      pages: Math.ceil(availableTasks.length / limitNum),
      hasNext: endIndex < availableTasks.length,
      hasPrev: pageNum > 0
    };
    
    logger.info(`获取可用任务: ${paginatedTasks.length}/${availableTasks.length} 条记录`);
    
    return res.paginated(paginatedTasks, pagination);
    
  } catch (error) {
    logger.error('获取可用任务异常:', error);
    return res.error(500, '获取可用任务失败');
  }
}

/**
 * 接受任务
 */
export async function acceptTask(req, res) {
  try {
    const { id } = req.params;
    
    const task = deliveryTasks.find(t => t.id === id);
    if (!task) {
      return res.error(404, '任务不存在');
    }
    
    if (task.status !== 'available') {
      return res.error(400, '任务已被其他信使接受');
    }
    
    // 检查信使是否存在且状态正常
    const courier = couriers.find(c => c.userId === req.user.id);
    if (!courier) {
      return res.error(403, '用户不是认证信使');
    }
    
    if (courier.status !== 'active') {
      return res.error(400, '信使状态异常，无法接受任务');
    }
    
    if (courier.currentTasks >= courier.maxTasks) {
      return res.error(400, '当前任务数量已达上限');
    }
    
    // 分配任务
    task.status = 'assigned';
    task.courierId = courier.id;
    task.courierInfo = {
      name: courier.profile.name,
      phone: courier.profile.phone,
      zone: courier.zone
    };
    task.updatedAt = new Date().toISOString();
    
    // 添加时间线记录
    task.timeline.push({
      status: 'assigned',
      timestamp: new Date().toISOString(),
      note: `任务被 ${courier.profile.name} 接受`,
      operator: courier.id
    });
    
    // 更新信使状态
    courier.currentTasks += 1;
    if (courier.currentTasks >= courier.maxTasks) {
      courier.status = 'busy';
    }
    
    logger.info(`任务接受: ${id} by ${courier.profile.name}`);
    
    return res.success(task, '任务接受成功');
    
  } catch (error) {
    logger.error('接受任务异常:', error);
    return res.error(500, '接受任务失败');
  }
}

/**
 * 更新任务状态
 */
export async function updateTaskStatus(req, res) {
  try {
    const { id } = req.params;
    const { status, note, location } = req.body;
    
    const task = deliveryTasks.find(t => t.id === id);
    if (!task) {
      return res.error(404, '任务不存在');
    }
    
    // 权限检查 - 只有任务分配的信使或管理员可以更新状态
    if (req.user.role !== 'super_admin' && 
        req.user.role !== 'admin' && 
        task.courierId !== req.user.id) {
      return res.error(403, '无权更新此任务状态');
    }
    
    // 验证状态转换的合法性
    const validTransitions = {
      available: ['assigned'],
      assigned: ['picked_up', 'available'], // 可以取消任务
      picked_up: ['in_transit'],
      in_transit: ['delivered'],
      delivered: ['completed']
    };
    
    if (!validTransitions[task.status] || !validTransitions[task.status].includes(status)) {
      return res.error(400, `无法从 ${task.status} 状态转换到 ${status} 状态`);
    }
    
    // 更新任务状态
    const oldStatus = task.status;
    task.status = status;
    task.updatedAt = new Date().toISOString();
    
    // 添加时间线记录
    task.timeline.push({
      status: status,
      timestamp: new Date().toISOString(),
      note: note || `状态从 ${oldStatus} 更新为 ${status}`,
      operator: req.user.id,
      location: location
    });
    
    // 如果任务完成，更新信使统计
    if (status === 'completed') {
      const courier = couriers.find(c => c.id === task.courierId);
      if (courier) {
        courier.completedTasks += 1;
        courier.currentTasks = Math.max(0, courier.currentTasks - 1);
        if (courier.status === 'busy' && courier.currentTasks < courier.maxTasks) {
          courier.status = 'active';
        }
      }
    }
    
    // 如果任务被取消，释放信使
    if (status === 'available' && oldStatus === 'assigned') {
      const courier = couriers.find(c => c.id === task.courierId);
      if (courier) {
        courier.currentTasks = Math.max(0, courier.currentTasks - 1);
        if (courier.status === 'busy' && courier.currentTasks < courier.maxTasks) {
          courier.status = 'active';
        }
      }
      task.courierId = null;
      task.courierInfo = null;
    }
    
    logger.info(`任务状态更新: ${id} ${oldStatus} -> ${status} by ${req.user.username}`);
    
    return res.success(task, '任务状态更新成功');
    
  } catch (error) {
    logger.error('更新任务状态异常:', error);
    return res.error(500, '更新任务状态失败');
  }
}

/**
 * 获取我的任务
 */
export async function getMyTasks(req, res) {
  try {
    const { page = 0, limit = 20, status } = req.query;
    
    // 查找信使信息
    const courier = couriers.find(c => c.userId === req.user.id);
    if (!courier) {
      return res.error(403, '用户不是认证信使');
    }
    
    let myTasks = deliveryTasks.filter(task => task.courierId === courier.id);
    
    // 状态过滤
    if (status) {
      myTasks = myTasks.filter(task => task.status === status);
    }
    
    // 按更新时间倒序排列
    myTasks.sort((a, b) => new Date(b.updatedAt) - new Date(a.updatedAt));
    
    // 分页处理
    const pageNum = parseInt(page, 10);
    const limitNum = parseInt(limit, 10);
    const startIndex = pageNum * limitNum;
    const endIndex = startIndex + limitNum;
    
    const paginatedTasks = myTasks.slice(startIndex, endIndex);
    
    const pagination = {
      page: pageNum,
      limit: limitNum,
      total: myTasks.length,
      pages: Math.ceil(myTasks.length / limitNum),
      hasNext: endIndex < myTasks.length,
      hasPrev: pageNum > 0
    };
    
    logger.info(`获取我的任务: ${paginatedTasks.length}/${myTasks.length} 条记录`);
    
    return res.paginated(paginatedTasks, pagination);
    
  } catch (error) {
    logger.error('获取我的任务异常:', error);
    return res.error(500, '获取我的任务失败');
  }
}

/**
 * 获取信使统计信息
 */
export async function getCourierStats(req, res) {
  try {
    const courier = couriers.find(c => c.userId === req.user.id);
    if (!courier) {
      return res.error(403, '用户不是认证信使');
    }
    
    const myTasks = deliveryTasks.filter(task => task.courierId === courier.id);
    
    const stats = {
      profile: courier.profile,
      status: courier.status,
      rating: courier.rating,
      completedTasks: courier.completedTasks,
      currentTasks: courier.currentTasks,
      maxTasks: courier.maxTasks,
      zone: courier.zone,
      taskStats: {
        total: myTasks.length,
        completed: myTasks.filter(t => t.status === 'completed').length,
        inProgress: myTasks.filter(t => ['assigned', 'picked_up', 'in_transit', 'delivered'].includes(t.status)).length,
        totalReward: myTasks.filter(t => t.status === 'completed').reduce((sum, t) => sum + t.reward, 0)
      },
      workingHours: courier.workingHours,
      lastLocationUpdate: courier.location.updateTime
    };
    
    logger.info(`获取信使统计: ${courier.profile.name}`);
    
    return res.success(stats, '获取统计信息成功');
    
  } catch (error) {
    logger.error('获取信使统计异常:', error);
    return res.error(500, '获取统计信息失败');
  }
}

/**
 * 申请成为信使
 */
export async function applyCourier(req, res) {
  try {
    const { zone, phone, idCard, experience } = req.body;
    
    // 验证输入
    const errors = [];
    if (!zone) errors.push({ field: 'zone', message: '工作区域不能为空' });
    if (!phone) errors.push({ field: 'phone', message: '联系电话不能为空' });
    if (!idCard) errors.push({ field: 'idCard', message: '身份证号不能为空' });
    
    if (errors.length > 0) {
      return res.validationError(errors);
    }
    
    // 检查是否已经是信使
    const existingCourier = couriers.find(c => c.userId === req.user.id);
    if (existingCourier) {
      return res.error(409, '您已经是认证信使');
    }
    
    // 在真实环境中，这里应该创建申请记录，等待审核
    // Mock 环境下直接创建信使
    const newCourier = {
      id: uuidv4(),
      userId: req.user.id,
      status: 'active',
      zone,
      rating: 5.0,
      completedTasks: 0,
      currentTasks: 0,
      maxTasks: 3,
      profile: {
        name: req.user.profile.fullName,
        phone,
        experience: experience || '新手信使',
        avatar: req.user.profile.avatar
      },
      location: {
        coordinates: { lat: 0, lng: 0 },
        updateTime: new Date().toISOString()
      },
      workingHours: {
        start: '08:00',
        end: '20:00',
        timezone: 'Asia/Shanghai'
      },
      applicationInfo: {
        idCard,
        experience,
        appliedAt: new Date().toISOString(),
        approvedAt: new Date().toISOString(),
        approvedBy: 'system'
      }
    };
    
    couriers.push(newCourier);
    
    logger.info(`信使申请成功: ${req.user.username} -> ${zone}`);
    
    return res.success(newCourier, '信使申请成功');
    
  } catch (error) {
    logger.error('信使申请异常:', error);
    return res.error(500, '信使申请失败');
  }
}

/**
 * 创建下级信使 - 高级信使可以创建下级信使
 */
export async function createCourier(req, res) {
  try {
    const { username, email, level, region, school, zone, building } = req.body;
    
    // 验证输入
    const errors = [];
    if (!username) errors.push({ field: 'username', message: '用户名不能为空' });
    if (!email) errors.push({ field: 'email', message: '邮箱不能为空' });
    if (!level) errors.push({ field: 'level', message: '信使级别不能为空' });
    
    if (errors.length > 0) {
      return res.validationError(errors);
    }
    
    // 获取当前用户的信使级别
    const currentCourier = couriers.find(c => c.userId === req.user.id);
    if (!currentCourier) {
      return res.error(403, '只有信使才能创建下级信使');
    }
    
    // 从用户角色中解析级别
    let currentLevel = 1;
    if (req.user.role === 'courier_level_4') currentLevel = 4;
    else if (req.user.role === 'courier_level_3') currentLevel = 3;
    else if (req.user.role === 'courier_level_2') currentLevel = 2;
    else if (req.user.role === 'courier_level_1') currentLevel = 1;
    
    // 权限验证：只能创建比自己低一级的信使
    if (level >= currentLevel) {
      return res.error(403, `${currentLevel}级信使只能创建${currentLevel-1}级或更低级别的信使`);
    }
    
    // 级别验证
    if (level < 1 || level > 4) {
      return res.error(400, '信使级别必须在1-4之间');
    }
    
    // 创建新信使信息
    const newCourier = {
      id: uuidv4(),
      userId: `user_${username}_${Date.now()}`,
      username,
      email, 
      level,
      status: 'pending', // 新创建的信使状态为待审核
      zone: zone || region || school || '默认区域',
      region: region || '默认区域',
      school: school || '默认学校',
      building: building || null,
      rating: 5.0,
      completedTasks: 0,
      currentTasks: 0,
      maxTasks: 3,
      parentId: req.user.id,
      parentCourierId: currentCourier.id,
      profile: {
        name: username,
        phone: '',
        experience: '新创建信使',
        avatar: null
      },
      location: {
        coordinates: { lat: 0, lng: 0 },
        updateTime: new Date().toISOString()
      },
      workingHours: {
        start: '08:00',
        end: '20:00',
        timezone: 'Asia/Shanghai'
      },
      createdAt: new Date().toISOString(),
      createdBy: req.user.id,
      createdByName: req.user.username
    };
    
    // 将新信使添加到列表
    couriers.push(newCourier);
    
    logger.info(`创建下级信使: ${username} (Level ${level}) by ${req.user.username} (Level ${currentLevel})`);
    
    return res.success(newCourier, '信使创建成功，等待审核');
    
  } catch (error) {
    logger.error('创建信使异常:', error);
    return res.error(500, '创建信使失败');
  }
}

/**
 * 获取可创建的信使级别 - 用于前端显示可选级别
 */
export async function getCreatableLevels(req, res) {
  try {
    // 获取当前用户的信使级别
    const currentCourier = couriers.find(c => c.userId === req.user.id);
    if (!currentCourier) {
      return res.error(403, '只有信使才能查看可创建级别');
    }
    
    // 从用户角色中解析级别
    let currentLevel = 1;
    if (req.user.role === 'courier_level_4') currentLevel = 4;
    else if (req.user.role === 'courier_level_3') currentLevel = 3;
    else if (req.user.role === 'courier_level_2') currentLevel = 2;
    else if (req.user.role === 'courier_level_1') currentLevel = 1;
    
    // 返回可创建的级别（比自己低一级）
    const creatableLevels = [];
    if (currentLevel > 1) {
      const targetLevel = currentLevel - 1;
      const levelNames = {
        1: '楼栋级信使',
        2: '片区级信使', 
        3: '学校级信使',
        4: '城市级信使'
      };
      
      creatableLevels.push({
        level: targetLevel,
        name: levelNames[targetLevel],
        description: `管理${targetLevel === 1 ? '楼栋' : targetLevel === 2 ? '片区' : targetLevel === 3 ? '学校' : '城市'}范围内的信件投递`
      });
    }
    
    const result = {
      current_level: currentLevel,
      creatable_levels: creatableLevels,
      can_create: creatableLevels.length > 0
    };
    
    logger.info(`获取可创建级别: ${req.user.username} (Level ${currentLevel})`);
    
    return res.success(result, '获取可创建级别成功');
    
  } catch (error) {
    logger.error('获取可创建级别异常:', error);
    return res.error(500, '获取可创建级别失败');
  }
}

/**
 * 获取下级信使列表 - 按层级权限显示
 * 4级信使可以看到1-3级所有信使，3级可以看到1-2级，以此类推
 */
export async function getSubordinateCouriersList(req, res) {
  try {
    const { level, page = 0, limit = 20, status, search } = req.query;
    
    // 获取当前用户级别
    let currentLevel = 1;
    if (req.user.role === 'courier_level_4') currentLevel = 4;
    else if (req.user.role === 'courier_level_3') currentLevel = 3;
    else if (req.user.role === 'courier_level_2') currentLevel = 2;
    else if (req.user.role === 'courier_level_1') currentLevel = 1;
    
    // 权限检查：只能查看比自己低级别的信使
    const viewableLevels = [];
    for (let i = 1; i < currentLevel; i++) {
      viewableLevels.push(i);
    }
    
    if (viewableLevels.length === 0) {
      return res.success({
        couriers: [],
        pagination: { page: 0, limit: parseInt(limit), total: 0 },
        viewable_levels: viewableLevels,
        current_level: currentLevel
      }, '当前级别无下级信使');
    }
    
    // 扩展现有信使数据，添加更多层级
    const allCouriers = [
      ...couriers,
      // 添加更多 mock 数据以展示层级
      {
        id: 'courier_l4_001',
        userId: 'user_l4_001',
        username: 'shanghai_city_manager',
        email: 'shanghai@openpenpal.com',
        level: 4,
        status: 'active',
        zone: '上海市',
        region: '上海市',
        school: '上海市高校',
        rating: 4.7,
        completedTasks: 678,
        currentTasks: 2,
        maxTasks: 5,
        parentId: null,
        profile: {
          name: '上海市总信使',
          phone: '13800138044',
          experience: '4年管理经验',
          avatar: '/avatars/courier_l4_001.png'
        },
        createdAt: '2023-05-01T00:00:00Z',
        createdBy: 'admin'
      },
      {
        id: 'courier_l3_001',
        userId: 'user_l3_001',
        username: 'tsinghua_school_manager',
        email: 'tsinghua@openpenpal.com',
        level: 3,
        status: 'active',
        zone: '清华大学',
        region: '清华大学',
        school: '清华大学',
        rating: 4.8,
        completedTasks: 345,
        currentTasks: 1,
        maxTasks: 3,
        parentId: 'courier_004',
        profile: {
          name: '清华大学总信使',
          phone: '13800138033',
          experience: '3年管理经验',
          avatar: '/avatars/courier_l3_001.png'
        },
        createdAt: '2023-06-15T00:00:00Z',
        createdBy: 'courier_004'
      },
      {
        id: 'courier_l2_001',
        userId: 'user_l2_001',
        username: 'pku_zone_a_manager',
        email: 'pku_zone_a@openpenpal.com',
        level: 2,
        status: 'active',
        zone: '北京大学A区',
        region: '北京大学A区',
        school: '北京大学',
        rating: 4.6,
        completedTasks: 156,
        currentTasks: 0,
        maxTasks: 2,
        parentId: 'courier_003',
        profile: {
          name: '北大A区信使',
          phone: '13800138022',
          experience: '2年配送经验',
          avatar: '/avatars/courier_l2_001.png'
        },
        createdAt: '2023-08-01T00:00:00Z',
        createdBy: 'courier_003'
      },
      {
        id: 'courier_l1_001',
        userId: 'user_l1_001',
        username: 'pku_building_1_manager',
        email: 'pku_b1@openpenpal.com',
        level: 1,
        status: 'active',
        zone: '北京大学1号楼',
        region: '北京大学A区',
        school: '北京大学',
        building: '1号楼',
        rating: 4.9,
        completedTasks: 89,
        currentTasks: 1,
        maxTasks: 1,
        parentId: 'courier_l2_001',
        profile: {
          name: '1号楼信使',
          phone: '13800138011',
          experience: '1年配送经验',
          avatar: '/avatars/courier_l1_001.png'
        },
        createdAt: '2023-09-01T00:00:00Z',
        createdBy: 'courier_l2_001'
      }
    ];
    
    // 筛选可查看的信使
    let filteredCouriers = allCouriers.filter(courier => {
      // 只显示比当前用户级别低的信使
      if (!viewableLevels.includes(courier.level)) return false;
      
      // 如果指定了级别过滤
      if (level && courier.level !== parseInt(level)) return false;
      
      // 状态过滤
      if (status && courier.status !== status) return false;
      
      // 搜索过滤
      if (search) {
        const searchLower = search.toLowerCase();
        return (
          courier.username.toLowerCase().includes(searchLower) ||
          courier.profile.name.toLowerCase().includes(searchLower) ||
          courier.zone.toLowerCase().includes(searchLower)
        );
      }
      
      return true;
    });
    
    // 按级别和创建时间排序
    filteredCouriers.sort((a, b) => {
      if (a.level !== b.level) return b.level - a.level; // 级别高的在前
      return new Date(b.createdAt) - new Date(a.createdAt); // 创建时间新的在前
    });
    
    // 分页处理
    const pageNum = parseInt(page);
    const limitNum = parseInt(limit);
    const startIndex = pageNum * limitNum;
    const endIndex = startIndex + limitNum;
    
    const paginatedCouriers = filteredCouriers.slice(startIndex, endIndex);
    
    const pagination = {
      page: pageNum,
      limit: limitNum,
      total: filteredCouriers.length,
      pages: Math.ceil(filteredCouriers.length / limitNum),
      hasNext: endIndex < filteredCouriers.length,
      hasPrev: pageNum > 0
    };
    
    logger.info(`获取下级信使列表: ${req.user.username} (Level ${currentLevel}) 查看 ${paginatedCouriers.length}/${filteredCouriers.length} 条记录`);
    
    return res.success({
      couriers: paginatedCouriers,
      pagination,
      viewable_levels: viewableLevels,
      current_level: currentLevel,
      level_statistics: {
        level_1: filteredCouriers.filter(c => c.level === 1).length,
        level_2: filteredCouriers.filter(c => c.level === 2).length,
        level_3: filteredCouriers.filter(c => c.level === 3).length,
      }
    }, '获取下级信使列表成功');
    
  } catch (error) {
    logger.error('获取下级信使列表异常:', error);
    return res.error(500, '获取下级信使列表失败');
  }
}

/**
 * 获取信使详细信息
 */
export async function getCourierDetails(req, res) {
  try {
    const { id } = req.params;
    
    // 获取当前用户级别
    let currentLevel = 1;
    if (req.user.role === 'courier_level_4') currentLevel = 4;
    else if (req.user.role === 'courier_level_3') currentLevel = 3;
    else if (req.user.role === 'courier_level_2') currentLevel = 2;
    else if (req.user.role === 'courier_level_1') currentLevel = 1;
    
    // 扩展信使列表查找目标信使
    const allCouriers = [
      ...couriers,
      {
        id: 'courier_l4_001',
        userId: 'user_l4_001',
        username: 'shanghai_city_manager',
        email: 'shanghai@openpenpal.com',
        level: 4,
        status: 'active',
        zone: '上海市',
        region: '上海市',
        school: '上海市高校',
        rating: 4.7,
        completedTasks: 678,
        currentTasks: 2,
        maxTasks: 5,
        parentId: null,
        profile: {
          name: '上海市总信使',
          phone: '13800138044',
          experience: '4年管理经验',
          avatar: '/avatars/courier_l4_001.png'
        },
        createdAt: '2023-05-01T00:00:00Z',
        createdBy: 'admin'
      }
    ];
    
    const courier = allCouriers.find(c => c.id === id);
    if (!courier) {
      return res.error(404, '信使不存在');
    }
    
    // 权限检查：只能查看比自己级别低的信使，或者自己
    if (courier.level >= currentLevel && courier.userId !== req.user.id) {
      return res.error(403, '无权查看此信使信息');
    }
    
    // 获取下级信使统计
    const subordinates = allCouriers.filter(c => c.parentId === courier.id);
    
    const courierDetails = {
      ...courier,
      subordinates_count: subordinates.length,
      subordinates_by_level: {
        level_1: subordinates.filter(c => c.level === 1).length,
        level_2: subordinates.filter(c => c.level === 2).length,
        level_3: subordinates.filter(c => c.level === 3).length,
      },
      recent_activities: [
        {
          id: 'activity_001',
          type: 'task_completed',
          description: '完成信件配送任务',
          timestamp: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString() // 2小时前
        },
        {
          id: 'activity_002',
          type: 'subordinate_created',
          description: '创建了新的下级信使',
          timestamp: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString() // 1天前
        }
      ],
      performance_metrics: {
        completion_rate: (courier.completedTasks / (courier.completedTasks + courier.currentTasks + 10)) * 100,
        average_response_time: 25, // 分钟
        satisfaction_score: courier.rating,
        monthly_tasks: 45
      }
    };
    
    logger.info(`获取信使详情: ${id} by ${req.user.username}`);
    
    return res.success(courierDetails, '获取信使详情成功');
    
  } catch (error) {
    logger.error('获取信使详情异常:', error);
    return res.error(500, '获取信使详情失败');
  }
}

// Mock 审批记录数据
let approvalRecords = [
  {
    id: 'approval_001',
    type: 'courier_modification',
    requesterId: 'courier_003',
    requesterName: '北京大学总信使',
    targetCourierId: 'courier_l2_001',
    targetCourierName: '北大A区信使',
    changes: {
      zone: { old: '北京大学A区', new: '北京大学B区' },
      phone: { old: '13800138022', new: '13800138099' }
    },
    reason: '工作区域调整',
    status: 'pending',
    approverId: 'courier_004',
    approverName: '北京市总信使',
    createdAt: new Date(Date.now() - 1 * 60 * 60 * 1000).toISOString(), // 1小时前
    reviewedAt: null
  }
];

/**
 * 修改信使信息 - 需要上级审批
 */
export async function modifyCourier(req, res) {
  try {
    const { id } = req.params;
    const { zone, phone, status, reason } = req.body;
    
    // 获取当前用户级别
    let currentLevel = 1;
    if (req.user.role === 'courier_level_4') currentLevel = 4;
    else if (req.user.role === 'courier_level_3') currentLevel = 3;
    else if (req.user.role === 'courier_level_2') currentLevel = 2;
    else if (req.user.role === 'courier_level_1') currentLevel = 1;
    
    // 查找目标信使
    const allCouriers = [
      ...couriers,
      {
        id: 'courier_l2_001',
        userId: 'user_l2_001',
        username: 'pku_zone_a_manager',
        email: 'pku_zone_a@openpenpal.com',
        level: 2,
        status: 'active',
        zone: '北京大学A区',
        region: '北京大学A区',
        school: '北京大学',
        rating: 4.6,
        completedTasks: 156,
        currentTasks: 0,
        maxTasks: 2,
        parentId: 'courier_003',
        profile: {
          name: '北大A区信使',
          phone: '13800138022',
          experience: '2年配送经验',
          avatar: '/avatars/courier_l2_001.png'
        },
        createdAt: '2023-08-01T00:00:00Z',
        createdBy: 'courier_003'
      }
    ];
    
    const targetCourier = allCouriers.find(c => c.id === id);
    if (!targetCourier) {
      return res.error(404, '信使不存在');
    }
    
    // 权限检查：只能修改比自己级别低的信使
    if (targetCourier.level >= currentLevel) {
      return res.error(403, '无权修改此级别信使信息');
    }
    
    // 验证修改原因
    if (!reason || reason.trim().length < 5) {
      return res.error(400, '修改原因不能少于5个字符');
    }
    
    // 收集变更内容
    const changes = {};
    if (zone && zone !== targetCourier.zone) {
      changes.zone = { old: targetCourier.zone, new: zone };
    }
    if (phone && phone !== targetCourier.profile.phone) {
      changes.phone = { old: targetCourier.profile.phone, new: phone };
    }
    if (status && status !== targetCourier.status) {
      changes.status = { old: targetCourier.status, new: status };
    }
    
    if (Object.keys(changes).length === 0) {
      return res.error(400, '没有检测到任何修改');
    }
    
    // 确定审批者（比当前用户高一级的信使）
    let approverId = null;
    let approverName = '系统管理员';
    
    if (currentLevel < 4) {
      // 查找上级信使
      const superiorCouriers = allCouriers.filter(c => c.level === currentLevel + 1);
      if (superiorCouriers.length > 0) {
        approverId = superiorCouriers[0].id;
        approverName = superiorCouriers[0].profile.name;
      }
    }
    
    // 创建审批记录
    const approvalRecord = {
      id: `approval_${Date.now()}`,
      type: 'courier_modification',
      requesterId: req.user.id,
      requesterName: req.user.profile?.fullName || req.user.username,
      targetCourierId: targetCourier.id,
      targetCourierName: targetCourier.profile.name,
      changes,
      reason: reason.trim(),
      status: 'pending',
      approverId,
      approverName,
      createdAt: new Date().toISOString(),
      reviewedAt: null
    };
    
    approvalRecords.push(approvalRecord);
    
    logger.info(`创建修改审批: ${targetCourier.profile.name} by ${req.user.username} -> ${approverName}`);
    
    return res.success({
      approval_record: approvalRecord,
      message: `修改申请已提交给 ${approverName}，等待审批`
    }, '修改申请提交成功');
    
  } catch (error) {
    logger.error('修改信使异常:', error);
    return res.error(500, '修改信使失败');
  }
}

/**
 * 获取待审批的修改申请列表
 */
export async function getPendingApprovals(req, res) {
  try {
    const { page = 0, limit = 20, status = 'pending' } = req.query;
    
    // 获取当前用户级别
    let currentLevel = 1;
    if (req.user.role === 'courier_level_4') currentLevel = 4;
    else if (req.user.role === 'courier_level_3') currentLevel = 3;
    else if (req.user.role === 'courier_level_2') currentLevel = 2;
    else if (req.user.role === 'courier_level_1') currentLevel = 1;
    
    // 筛选需要当前用户审批的记录
    let filteredRecords = approvalRecords.filter(record => {
      // 只显示指定给当前用户审批的记录
      if (record.approverId !== req.user.id && currentLevel < 4) return false;
      
      // 4级信使可以看到所有审批记录
      if (currentLevel === 4) return true;
      
      // 状态过滤
      if (status && record.status !== status) return false;
      
      return true;
    });
    
    // 按创建时间倒序排列
    filteredRecords.sort((a, b) => new Date(b.createdAt) - new Date(a.createdAt));
    
    // 分页处理
    const pageNum = parseInt(page);
    const limitNum = parseInt(limit);
    const startIndex = pageNum * limitNum;
    const endIndex = startIndex + limitNum;
    
    const paginatedRecords = filteredRecords.slice(startIndex, endIndex);
    
    const pagination = {
      page: pageNum,
      limit: limitNum,
      total: filteredRecords.length,
      pages: Math.ceil(filteredRecords.length / limitNum),
      hasNext: endIndex < filteredRecords.length,
      hasPrev: pageNum > 0
    };
    
    logger.info(`获取待审批列表: ${req.user.username} (Level ${currentLevel}) 查看 ${paginatedRecords.length}/${filteredRecords.length} 条记录`);
    
    return res.success({
      approvals: paginatedRecords,
      pagination,
      statistics: {
        pending: filteredRecords.filter(r => r.status === 'pending').length,
        approved: filteredRecords.filter(r => r.status === 'approved').length,
        rejected: filteredRecords.filter(r => r.status === 'rejected').length
      }
    }, '获取待审批列表成功');
    
  } catch (error) {
    logger.error('获取待审批列表异常:', error);
    return res.error(500, '获取待审批列表失败');
  }
}

/**
 * 审批修改申请
 */
export async function approveModification(req, res) {
  try {
    const { id } = req.params;
    const { action, comment } = req.body; // action: 'approve' | 'reject'
    
    // 验证输入
    if (!['approve', 'reject'].includes(action)) {
      return res.error(400, '审批动作必须是 approve 或 reject');
    }
    
    // 查找审批记录
    const approvalRecord = approvalRecords.find(r => r.id === id);
    if (!approvalRecord) {
      return res.error(404, '审批记录不存在');
    }
    
    if (approvalRecord.status !== 'pending') {
      return res.error(400, '此申请已经被处理过了');
    }
    
    // 权限检查：只有指定的审批者才能处理
    if (approvalRecord.approverId !== req.user.id && req.user.role !== 'super_admin') {
      return res.error(403, '无权处理此审批申请');
    }
    
    // 更新审批记录
    approvalRecord.status = action === 'approve' ? 'approved' : 'rejected';
    approvalRecord.reviewedAt = new Date().toISOString();
    approvalRecord.reviewComment = comment || '';
    approvalRecord.reviewedBy = req.user.username;
    
    // 如果是批准，应用修改到目标信使（在实际系统中会更新数据库）
    if (action === 'approve') {
      logger.info(`审批通过，应用修改到信使: ${approvalRecord.targetCourierName}`);
      // 这里应该实际更新信使信息，但在mock系统中我们只记录日志
    }
    
    logger.info(`审批处理: ${approvalRecord.id} ${action} by ${req.user.username}`);
    
    return res.success({
      approval_record: approvalRecord,
      message: action === 'approve' ? '审批通过，修改已生效' : '审批拒绝'
    }, '审批处理成功');
    
  } catch (error) {
    logger.error('审批处理异常:', error);
    return res.error(500, '审批处理失败');
  }
}

/**
 * 获取我提交的修改申请
 */
export async function getMyModificationRequests(req, res) {
  try {
    const { page = 0, limit = 20, status } = req.query;
    
    // 筛选当前用户提交的申请
    let myRequests = approvalRecords.filter(record => record.requesterId === req.user.id);
    
    // 状态过滤
    if (status) {
      myRequests = myRequests.filter(record => record.status === status);
    }
    
    // 按创建时间倒序排列
    myRequests.sort((a, b) => new Date(b.createdAt) - new Date(a.createdAt));
    
    // 分页处理
    const pageNum = parseInt(page);
    const limitNum = parseInt(limit);
    const startIndex = pageNum * limitNum;
    const endIndex = startIndex + limitNum;
    
    const paginatedRequests = myRequests.slice(startIndex, endIndex);
    
    const pagination = {
      page: pageNum,
      limit: limitNum,
      total: myRequests.length,
      pages: Math.ceil(myRequests.length / limitNum),
      hasNext: endIndex < myRequests.length,
      hasPrev: pageNum > 0
    };
    
    logger.info(`获取我的修改申请: ${req.user.username} 查看 ${paginatedRequests.length}/${myRequests.length} 条记录`);
    
    return res.success({
      requests: paginatedRequests,
      pagination,
      statistics: {
        pending: myRequests.filter(r => r.status === 'pending').length,
        approved: myRequests.filter(r => r.status === 'approved').length,
        rejected: myRequests.filter(r => r.status === 'rejected').length
      }
    }, '获取我的修改申请成功');
    
  } catch (error) {
    logger.error('获取我的修改申请异常:', error);
    return res.error(500, '获取我的修改申请失败');
  }
}