// 层级信使测试账号配置
// 这些账号用于演示四级信使管理体系

export interface CourierTestAccount {
  username: string
  password: string
  email: string
  level: number
  levelName: string
  zoneCode: string
  zoneType: 'city' | 'school' | 'zone' | 'building'
  parentUsername?: string
  description: string
  permissions: string[]
  managementPath: string
}

export const COURIER_TEST_ACCOUNTS: CourierTestAccount[] = [
  {
    username: 'courier_level4_city',
    password: 'city123',
    email: 'city.courier@openpenpal.com',
    level: 4,
    levelName: '四级信使（城市总代）',
    zoneCode: 'BEIJING',
    zoneType: 'city',
    description: '北京市信使总负责人，管理全市学校信使网络',
    permissions: [
      'courier_scan_code',
      'courier_deliver_letter',
      'courier_view_own_tasks',
      'courier_report_exception',
      'courier_manage_subordinates',
      'courier_assign_tasks',
      'courier_view_subordinate_reports',
      'courier_create_subordinate',
      'courier_manage_school_zone',
      'courier_view_school_analytics',
      'courier_coordinate_cross_zone',
      'courier_manage_city_operations',
      'courier_create_school_courier',
      'courier_view_city_analytics',
    ],
    managementPath: '/courier/city-manage',
  },
  {
    username: 'courier_level3_school',
    password: 'school123',
    email: 'school.courier@openpenpal.com',
    level: 3,
    levelName: '三级信使（校级）',
    zoneCode: 'BJDX',
    zoneType: 'school',
    parentUsername: 'courier_level4_city',
    description: '北京大学信使负责人，管理全校信使团队',
    permissions: [
      'courier_scan_code',
      'courier_deliver_letter',
      'courier_view_own_tasks',
      'courier_report_exception',
      'courier_manage_subordinates',
      'courier_assign_tasks',
      'courier_view_subordinate_reports',
      'courier_create_subordinate',
      'courier_manage_school_zone',
      'courier_view_school_analytics',
      'courier_coordinate_cross_zone',
    ],
    managementPath: '/courier/school-manage',
  },
  {
    username: 'courier_level2_zone',
    password: 'zone123',
    email: 'zone.courier@openpenpal.com',
    level: 2,
    levelName: '二级信使（片区/年级）',
    zoneCode: 'BJDX_ZONE_A',
    zoneType: 'zone',
    parentUsername: 'courier_level3_school',
    description: '北大A区信使组长，管理片区内一级信使',
    permissions: [
      'courier_scan_code',
      'courier_deliver_letter',
      'courier_view_own_tasks',
      'courier_report_exception',
      'courier_manage_subordinates',
      'courier_assign_tasks',
      'courier_view_subordinate_reports',
      'courier_create_subordinate',
    ],
    managementPath: '/courier/zone-manage',
  },
  {
    username: 'courier_level1_basic',
    password: 'basic123',
    email: 'basic.courier@openpenpal.com',
    level: 1,
    levelName: '一级信使（楼栋/班级）',
    zoneCode: 'BJDX_BUILDING_32',
    zoneType: 'building',
    parentUsername: 'courier_level2_zone',
    description: '32号楼信使，负责楼栋内信件收发',
    permissions: [
      'courier_scan_code',
      'courier_deliver_letter',
      'courier_view_own_tasks',
      'courier_report_exception',
    ],
    managementPath: '/courier/tasks',
  },
  // 额外的测试账号，用于测试多个下级
  {
    username: 'courier_level1_basic2',
    password: 'basic123',
    email: 'basic2.courier@openpenpal.com',
    level: 1,
    levelName: '一级信使（楼栋/班级）',
    zoneCode: 'BJDX_BUILDING_33',
    zoneType: 'building',
    parentUsername: 'courier_level2_zone',
    description: '33号楼信使，负责楼栋内信件收发',
    permissions: [
      'courier_scan_code',
      'courier_deliver_letter',
      'courier_view_own_tasks',
      'courier_report_exception',
    ],
    managementPath: '/courier/tasks',
  },
  {
    username: 'courier_level2_zone_b',
    password: 'zone123',
    email: 'zone.b.courier@openpenpal.com',
    level: 2,
    levelName: '二级信使（片区/年级）',
    zoneCode: 'BJDX_ZONE_B',
    zoneType: 'zone',
    parentUsername: 'courier_level3_school',
    description: '北大B区信使组长，管理片区内一级信使',
    permissions: [
      'courier_scan_code',
      'courier_deliver_letter',
      'courier_view_own_tasks',
      'courier_report_exception',
      'courier_manage_subordinates',
      'courier_assign_tasks',
      'courier_view_subordinate_reports',
      'courier_create_subordinate',
    ],
    managementPath: '/courier/zone-manage',
  },
]

// 获取特定级别的测试账号
export function getCourierTestAccountsByLevel(level: number): CourierTestAccount[] {
  return COURIER_TEST_ACCOUNTS.filter(account => account.level === level)
}

// 获取特定账号的下级账号
export function getSubordinateAccounts(username: string): CourierTestAccount[] {
  return COURIER_TEST_ACCOUNTS.filter(account => account.parentUsername === username)
}

// 生成测试账号的模拟登录数据
export function generateCourierMockData(account: CourierTestAccount) {
  // 根据级别映射到正确的角色
  const roleMapping: Record<number, string> = {
    1: 'courier_level1',
    2: 'courier_level2', 
    3: 'courier_level3',
    4: 'courier_level4'
  }
  
  return {
    id: `courier_${account.username}`,
    username: account.username,
    email: account.email,
    role: roleMapping[account.level] || 'courier_level1', // 根据级别设置角色
    courierInfo: {
      level: account.level,
      zoneCode: account.zoneCode,
      zoneType: account.zoneType,
      status: 'active',
      points: Math.floor(Math.random() * 1000) + 100,
      taskCount: Math.floor(Math.random() * 50) + 10,
    },
    permissions: account.permissions,
  }
}