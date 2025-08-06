import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    name: 'Layout',
    component: () => import('@/layouts/AdminLayout.vue'),
    meta: { requiresAuth: true },
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '仪表板', icon: 'DataAnalysis' }
      },
      {
        path: 'users',
        name: 'Users',
        component: () => import('@/views/Users.vue'),
        meta: { title: '用户管理', icon: 'User', permission: 'user.read' }
      },
      {
        path: 'letters',
        name: 'Letters',
        component: () => import('@/views/Letters.vue'),
        meta: { title: '信件管理', icon: 'Message', permission: 'letter.read' }
      },
      {
        path: 'couriers',
        name: 'Couriers',
        component: () => import('@/views/Couriers.vue'),
        meta: { title: '信使管理', icon: 'TruckFilled', permission: 'courier.read' }
      },
      {
        path: 'museum',
        name: 'Museum',
        component: () => import('@/views/Museum.vue'),
        meta: { title: '信件博物馆', icon: 'CollectionTag', permission: 'museum.read' }
      },
      {
        path: 'statistics',
        name: 'Statistics',
        component: () => import('@/views/Statistics.vue'),
        meta: { title: '数据统计', icon: 'PieChart', permission: 'stats.read' }
      },
      {
        path: 'settings',
        name: 'Settings',
        component: () => import('@/views/Settings.vue'),
        meta: { title: '系统设置', icon: 'Setting', permission: 'config.read' }
      }
    ]
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/NotFound.vue')
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()
  
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next('/login')
  } else if (to.path === '/login' && authStore.isAuthenticated) {
    next('/')
  } else {
    next()
  }
})

export default router