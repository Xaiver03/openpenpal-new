<template>
  <el-container class="admin-layout">
    <!-- 侧边栏 -->
    <el-aside :width="isCollapse ? '64px' : '256px'" class="sidebar">
      <div class="logo">
        <div v-if="!isCollapse" class="logo-text">OpenPenPal</div>
        <div v-else class="logo-mini">OP</div>
      </div>
      
      <el-menu
        :default-active="$route.path"
        class="sidebar-menu"
        :collapse="isCollapse"
        router
        background-color="#304156"
        text-color="#bfcbd9"
        active-text-color="#409EFF"
      >
        <template v-for="route in menuRoutes" :key="route.path">
          <el-menu-item 
            v-if="hasPermission(route.meta?.permission)"
            :index="route.path"
          >
            <el-icon>
              <component :is="route.meta?.icon" />
            </el-icon>
            <template #title>{{ route.meta?.title }}</template>
          </el-menu-item>
        </template>
      </el-menu>
    </el-aside>

    <!-- 主要内容区域 -->
    <el-container>
      <!-- 顶部导航栏 -->
      <el-header class="header">
        <div class="header-left">
          <el-button
            type="text"
            @click="toggleSidebar"
            class="sidebar-toggle"
          >
            <el-icon>
              <Fold v-if="!isCollapse" />
              <Expand v-else />
            </el-icon>
          </el-button>
          
          <el-breadcrumb separator="/">
            <el-breadcrumb-item>管理后台</el-breadcrumb-item>
            <el-breadcrumb-item>{{ currentPageTitle }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>

        <div class="header-right">
          <!-- 通知图标 -->
          <el-badge :value="notificationCount" class="notification">
            <el-icon size="20">
              <Bell />
            </el-icon>
          </el-badge>

          <!-- 用户菜单 -->
          <el-dropdown @command="handleUserCommand">
            <div class="user-info">
              <el-avatar :src="authStore.user?.avatarUrl" size="small">
                {{ authStore.user?.username?.charAt(0).toUpperCase() }}
              </el-avatar>
              <span class="username">{{ authStore.user?.username }}</span>
              <el-icon>
                <ArrowDown />
              </el-icon>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">个人资料</el-dropdown-item>
                <el-dropdown-item command="settings">系统设置</el-dropdown-item>
                <el-dropdown-item divided command="logout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <!-- 主内容区域 -->
      <el-main class="main-content">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Fold,
  Expand,
  Bell,
  ArrowDown,
  DataAnalysis,
  User,
  Message,
  Van,
  PieChart,
  Setting
} from '@element-plus/icons-vue'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

// 侧边栏折叠状态
const isCollapse = ref(false)

// 通知数量（模拟）
const notificationCount = ref(3)

// 菜单路由配置
const menuRoutes = [
  {
    path: '/dashboard',
    meta: { title: '仪表板', icon: 'DataAnalysis' }
  },
  {
    path: '/users',
    meta: { title: '用户管理', icon: 'User', permission: 'user.read' }
  },
  {
    path: '/letters',
    meta: { title: '信件管理', icon: 'Message', permission: 'letter.read' }
  },
  {
    path: '/couriers',
    meta: { title: '信使管理', icon: 'Van', permission: 'courier.read' }
  },
  {
    path: '/statistics',
    meta: { title: '数据统计', icon: 'PieChart', permission: 'stats.read' }
  },
  {
    path: '/settings',
    meta: { title: '系统设置', icon: 'Setting', permission: 'config.read' }
  }
]

// 当前页面标题
const currentPageTitle = computed(() => {
  const currentRoute = menuRoutes.find(r => r.path === route.path)
  return currentRoute?.meta?.title || '未知页面'
})

// 权限检查
const hasPermission = (permission?: string) => {
  if (!permission) return true
  return authStore.hasPermission(permission)
}

// 切换侧边栏
const toggleSidebar = () => {
  isCollapse.value = !isCollapse.value
}

// 处理用户菜单命令
const handleUserCommand = async (command: string) => {
  switch (command) {
    case 'profile':
      // 跳转到个人资料页面
      ElMessage.info('个人资料功能开发中...')
      break
    case 'settings':
      router.push('/settings')
      break
    case 'logout':
      try {
        await ElMessageBox.confirm('确定要退出登录吗？', '确认退出', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        })
        authStore.logout()
        router.push('/login')
        ElMessage.success('已退出登录')
      } catch {
        // 用户取消操作
      }
      break
  }
}
</script>

<style scoped>
.admin-layout {
  height: 100vh;
}

.sidebar {
  background-color: #304156;
  transition: width 0.3s;
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #2b3a4b;
}

.logo img {
  height: 32px;
}

.sidebar-menu {
  border: none;
}

.sidebar-menu:not(.el-menu--collapse) {
  width: 256px;
}

.header {
  background-color: #fff;
  border-bottom: 1px solid #e6e6e6;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 20px;
}

.sidebar-toggle {
  font-size: 20px;
  color: #666;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 20px;
}

.notification {
  cursor: pointer;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 8px;
  border-radius: 4px;
  transition: background-color 0.3s;
}

.user-info:hover {
  background-color: #f5f5f5;
}

.username {
  font-size: 14px;
  color: #333;
}

.main-content {
  background-color: #f5f5f5;
  padding: 20px;
}
</style>