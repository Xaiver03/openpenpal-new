// 用户相关类型
export interface User {
  id: string
  username: string
  email: string
  role: string
  schoolCode?: string
  status: 'ACTIVE' | 'INACTIVE' | 'SUSPENDED' | 'BANNED'
  lastLogin?: string
  failedLoginAttempts: number
  lockedUntil?: string
  avatarUrl?: string
  bio?: string
  permissions: string[]
  createdAt: string
  updatedAt: string
  statistics?: UserStatistics
}

export interface UserStatistics {
  lettersSent: number
  lettersReceived: number
  courierTasks: number
}

export interface LoginForm {
  username: string
  password: string
}

// 信件相关类型
export interface Letter {
  id: string
  title: string
  content?: string
  status: 'draft' | 'generated' | 'collected' | 'in_transit' | 'delivered' | 'failed'
  urgent: boolean
  createdAt: string
  updatedAt: string
  sender: {
    id: string
    username: string
    schoolCode: string
  }
  courier?: {
    id: string
    username: string
  }
  receiverHint: string
  qrCode?: string
}

// 信使相关类型
export interface Courier {
  id: string
  user: {
    id: string
    username: string
    email: string
    schoolCode: string
  }
  zone: string
  status: 'pending' | 'approved' | 'active' | 'suspended' | 'banned'
  rating: number
  currentTasks: number
  lastActive: string
  createdAt: string
  updatedAt: string
  statistics: CourierStatistics
}

export interface CourierStatistics {
  totalDeliveries: number
  successfulDeliveries: number
  failedDeliveries: number
  averageDeliveryTime: number
  successRate: number
}

// 统计相关类型
export interface DashboardStats {
  userStats: {
    total: number
    active: number
    byRole: Record<string, number>
    bySchool: Record<string, number>
  }
  letterStats: {
    total: number
    byStatus: Record<string, number>
    todayCount: number
    weekCount: number
  }
  courierStats: {
    total: number
    active: number
    averageRating: number
    averageDeliveryTime: number
  }
}

// 表格查询参数
export interface TableQuery {
  page: number
  size: number
  search?: string
  sort?: string
  direction?: 'asc' | 'desc'
  [key: string]: any
}

// 菜单项类型
export interface MenuItem {
  path: string
  name: string
  title: string
  icon?: string
  permission?: string
  children?: MenuItem[]
}