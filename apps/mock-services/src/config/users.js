/**
 * Mock 用户和权限配置
 * 统一管理所有测试用户的账号、角色、权限信息
 */

export const USERS = {
  // 超级管理员
  admin: {
    id: 'admin_001',
    username: 'admin',
    password: 'admin123',
    email: 'admin@openpenpal.com',
    role: 'super_admin',
    schoolCode: 'ADMIN',
    permissions: [
      'ALL', // 超级权限，可访问所有服务
      'ADMIN_READ', 'ADMIN_WRITE', 'ADMIN_DELETE',
      'USER_MANAGE', 'SYSTEM_CONFIG',
      'CONTENT_MODERATE', 'MUSEUM_MANAGE'
    ],
    profile: {
      fullName: '系统管理员',
      avatar: '/avatars/admin.png'
    }
  },

  // 普通学生用户
  alice: {
    id: 'user_001',
    username: 'alice',
    password: 'secret',
    email: 'alice@pku.edu.cn',
    role: 'student',
    schoolCode: 'PKU',
    permissions: [
      'WRITE_READ', 'WRITE_CREATE',
      'LETTER_READ', 'LETTER_SEND',
      'PROFILE_READ', 'PROFILE_UPDATE'
    ],
    profile: {
      fullName: '爱丽丝',
      grade: '大二',
      major: '计算机科学',
      avatar: '/avatars/alice.png'
    }
  },

  // 学生用户2
  bob: {
    id: 'user_002',
    username: 'bob',
    password: 'password123',
    email: 'bob@tsinghua.edu.cn',
    role: 'student',
    schoolCode: 'THU',
    permissions: [
      'WRITE_READ', 'WRITE_CREATE',
      'LETTER_READ', 'LETTER_SEND',
      'PROFILE_READ', 'PROFILE_UPDATE'
    ],
    profile: {
      fullName: '鲍勃',
      grade: '大三',
      major: '物理学',
      avatar: '/avatars/bob.png'
    }
  },

  // 信使用户1
  courier1: {
    id: 'courier_001',
    username: 'courier1',
    password: 'courier123',
    email: 'courier1@openpenpal.com',
    role: 'courier',
    schoolCode: 'PKU',
    permissions: [
      'COURIER_READ', 'COURIER_WRITE',
      'TASK_READ', 'TASK_ACCEPT', 'TASK_COMPLETE',
      'DELIVERY_MANAGE', 'ROUTE_PLAN'
    ],
    profile: {
      fullName: '快递员小王',
      phone: '13800138001',
      zone: '北京大学',
      avatar: '/avatars/courier1.png'
    },
    courierInfo: {
      status: 'active',
      rating: 4.8,
      completedTasks: 156,
      experience: '2年配送经验'
    }
  },

  // 信使用户2
  courier2: {
    id: 'courier_002',
    username: 'courier2',
    password: 'courier456',
    email: 'courier2@openpenpal.com',
    role: 'courier',
    schoolCode: 'THU',
    permissions: [
      'COURIER_READ', 'COURIER_WRITE',
      'TASK_READ', 'TASK_ACCEPT', 'TASK_COMPLETE',
      'DELIVERY_MANAGE', 'ROUTE_PLAN'
    ],
    profile: {
      fullName: '快递员小李',
      phone: '13800138002',
      zone: '清华大学',
      avatar: '/avatars/courier2.png'
    },
    courierInfo: {
      status: 'active',
      rating: 4.9,
      completedTasks: 203,
      experience: '3年配送经验'
    }
  },

  // 教师用户
  teacher1: {
    id: 'teacher_001',
    username: 'teacher1',
    password: 'teacher123',
    email: 'teacher1@pku.edu.cn',
    role: 'teacher',
    schoolCode: 'PKU',
    permissions: [
      'WRITE_READ', 'WRITE_CREATE',
      'LETTER_READ', 'LETTER_SEND',
      'PROFILE_READ', 'PROFILE_UPDATE',
      'STUDENT_MANAGE', 'CLASS_MANAGE'
    ],
    profile: {
      fullName: '张教授',
      department: '计算机学院',
      title: '副教授',
      avatar: '/avatars/teacher1.png'
    }
  },

  // 四级信使（城市级）
  courier4: {
    id: 'courier_004',
    username: 'courier4',
    password: 'courier123',
    email: 'courier4@openpenpal.com',
    role: 'courier_level_4',
    schoolCode: 'BEIJING_CITY',
    permissions: [
      'COURIER_READ', 'COURIER_WRITE', 'COURIER_MANAGE',
      'TASK_READ', 'TASK_ACCEPT', 'TASK_COMPLETE',
      'DELIVERY_MANAGE', 'ROUTE_PLAN',
      'SUBORDINATE_MANAGE', 'CREATE_LOWER_LEVEL_COURIER'
    ],
    profile: {
      fullName: '北京市总信使',
      phone: '13800138004',
      zone: '北京市',
      avatar: '/avatars/courier4.png'
    },
    courierInfo: {
      level: 4,
      status: 'active',
      rating: 4.9,
      completedTasks: 856,
      experience: '5年管理经验'
    }
  },

  // 三级信使（学校级）
  courier3: {
    id: 'courier_003',
    username: 'courier3',
    password: 'courier123',
    email: 'courier3@openpenpal.com',
    role: 'courier_level_3',
    schoolCode: 'PKU',
    permissions: [
      'COURIER_READ', 'COURIER_WRITE', 'COURIER_MANAGE',
      'TASK_READ', 'TASK_ACCEPT', 'TASK_COMPLETE',
      'DELIVERY_MANAGE', 'ROUTE_PLAN',
      'SUBORDINATE_MANAGE'
    ],
    profile: {
      fullName: '北京大学总信使',
      phone: '13800138003',
      zone: '北京大学',
      avatar: '/avatars/courier3.png'
    },
    courierInfo: {
      level: 3,
      status: 'active',
      rating: 4.8,
      completedTasks: 456,
      experience: '3年管理经验'
    }
  },

  // 审核员
  moderator: {
    id: 'mod_001',
    username: 'moderator',
    password: 'mod123',
    email: 'moderator@openpenpal.com',
    role: 'moderator',
    schoolCode: 'SYSTEM',
    permissions: [
      'CONTENT_READ', 'CONTENT_MODERATE',
      'MUSEUM_READ', 'MUSEUM_MODERATE',
      'REPORT_HANDLE', 'SENSITIVE_WORD_MANAGE'
    ],
    profile: {
      fullName: '内容审核员',
      avatar: '/avatars/moderator.png'
    }
  }
};

// 角色权限映射
export const ROLE_PERMISSIONS = {
  super_admin: ['ALL'],
  admin: [
    'ADMIN_READ', 'ADMIN_WRITE', 'ADMIN_DELETE',
    'USER_MANAGE', 'SYSTEM_CONFIG',
    'CONTENT_MODERATE', 'MUSEUM_MANAGE'
  ],
  student: [
    'WRITE_READ', 'WRITE_CREATE',
    'LETTER_READ', 'LETTER_SEND',
    'PROFILE_READ', 'PROFILE_UPDATE'
  ],
  teacher: [
    'WRITE_READ', 'WRITE_CREATE',
    'LETTER_READ', 'LETTER_SEND',
    'PROFILE_READ', 'PROFILE_UPDATE',
    'STUDENT_MANAGE', 'CLASS_MANAGE'
  ],
  courier: [
    'COURIER_READ', 'COURIER_WRITE',
    'TASK_READ', 'TASK_ACCEPT', 'TASK_COMPLETE',
    'DELIVERY_MANAGE', 'ROUTE_PLAN'
  ],
  courier_level_1: [
    'COURIER_READ', 'COURIER_WRITE',
    'TASK_READ', 'TASK_ACCEPT', 'TASK_COMPLETE',
    'DELIVERY_MANAGE', 'ROUTE_PLAN'
  ],
  courier_level_2: [
    'COURIER_READ', 'COURIER_WRITE',
    'TASK_READ', 'TASK_ACCEPT', 'TASK_COMPLETE',
    'DELIVERY_MANAGE', 'ROUTE_PLAN',
    'SUBORDINATE_MANAGE'
  ],
  courier_level_3: [
    'COURIER_READ', 'COURIER_WRITE', 'COURIER_MANAGE',
    'TASK_READ', 'TASK_ACCEPT', 'TASK_COMPLETE',
    'DELIVERY_MANAGE', 'ROUTE_PLAN',
    'SUBORDINATE_MANAGE'
  ],
  courier_level_4: [
    'COURIER_READ', 'COURIER_WRITE', 'COURIER_MANAGE',
    'TASK_READ', 'TASK_ACCEPT', 'TASK_COMPLETE',
    'DELIVERY_MANAGE', 'ROUTE_PLAN',
    'SUBORDINATE_MANAGE', 'CREATE_LOWER_LEVEL_COURIER'
  ],
  moderator: [
    'CONTENT_READ', 'CONTENT_MODERATE',
    'MUSEUM_READ', 'MUSEUM_MODERATE',
    'REPORT_HANDLE', 'SENSITIVE_WORD_MANAGE'
  ]
};

// 服务权限映射
export const SERVICE_PERMISSIONS = {
  'write-service': ['WRITE_READ', 'WRITE_CREATE', 'LETTER_READ', 'LETTER_SEND'],
  'courier-service': ['COURIER_READ', 'COURIER_WRITE', 'TASK_READ', 'TASK_ACCEPT', 'TASK_COMPLETE'],
  'admin-service': ['ADMIN_READ', 'ADMIN_WRITE', 'USER_MANAGE', 'SYSTEM_CONFIG'],
  'main-backend': ['PROFILE_READ', 'PROFILE_UPDATE', 'USER_MANAGE'],
  'ocr-service': ['OCR_READ', 'OCR_PROCESS']
};

// 根据用户名查找用户
export function findUserByUsername(username) {
  return USERS[username] || null;
}

// 根据用户ID查找用户
export function findUserById(id) {
  return Object.values(USERS).find(user => user.id === id) || null;
}

// 检查用户是否有指定权限
export function hasPermission(user, permission) {
  if (!user || !user.permissions) return false;
  
  // 超级管理员有所有权限
  if (user.permissions.includes('ALL')) return true;
  
  return user.permissions.includes(permission);
}

// 检查用户是否可以访问指定服务
export function canAccessService(user, serviceName) {
  if (!user || !user.permissions) return false;
  
  // 超级管理员可以访问所有服务
  if (user.permissions.includes('ALL')) return true;
  
  const requiredPermissions = SERVICE_PERMISSIONS[serviceName] || [];
  return requiredPermissions.some(permission => user.permissions.includes(permission));
}