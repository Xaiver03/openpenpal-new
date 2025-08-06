<template>
  <div class="users-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <h1>用户管理</h1>
      <p>管理系统中的所有用户账户</p>
    </div>

    <!-- 搜索和过滤 -->
    <el-card class="filter-card">
      <el-form :model="searchForm" inline>
        <el-form-item label="搜索">
          <el-input
            v-model="searchForm.search"
            placeholder="输入用户名或邮箱"
            clearable
            style="width: 200px"
            @keyup.enter="handleSearch"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        
        <el-form-item label="角色">
          <el-select v-model="searchForm.role" placeholder="选择角色" clearable style="width: 150px">
            <el-option label="普通用户" value="user" />
            <el-option label="信使" value="courier" />
            <el-option label="管理员" value="admin" />
            <el-option label="超级管理员" value="super_admin" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="选择状态" clearable style="width: 120px">
            <el-option label="激活" value="ACTIVE" />
            <el-option label="禁用" value="INACTIVE" />
            <el-option label="锁定" value="LOCKED" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="学校">
          <el-input
            v-model="searchForm.schoolCode"
            placeholder="学校代码"
            clearable
            style="width: 150px"
          />
        </el-form-item>
        
        <el-form-item>
          <el-button type="primary" @click="handleSearch" :loading="loading">
            <el-icon><Search /></el-icon>
            搜索
          </el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 用户列表 -->
    <el-card>
      <template #header>
        <div class="table-header">
          <span>用户列表 ({{ pagination.total }})</span>
          <div class="table-actions">
            <el-button type="success" size="small" @click="exportUsers">
              <el-icon><Download /></el-icon>
              导出
            </el-button>
          </div>
        </div>
      </template>

      <el-table
        :data="users"
        :loading="loading"
        stripe
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        
        <el-table-column label="用户信息" min-width="200">
          <template #default="scope">
            <div class="user-info">
              <el-avatar :src="scope.row.avatarUrl" size="small">
                {{ scope.row.username.charAt(0).toUpperCase() }}
              </el-avatar>
              <div class="user-details">
                <div class="username">{{ scope.row.username }}</div>
                <div class="email">{{ scope.row.email }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column label="角色" width="120">
          <template #default="scope">
            <el-tag :type="getRoleTagType(scope.row.role)" size="small">
              {{ getRoleText(scope.row.role) }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column label="状态" width="100">
          <template #default="scope">
            <el-tag :type="getStatusTagType(scope.row.status)" size="small">
              {{ getStatusText(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column label="学校" prop="schoolCode" width="120" />
        
        <el-table-column label="统计信息" width="150">
          <template #default="scope">
            <div class="user-stats">
              <div>信件: {{ scope.row.statistics?.lettersSent || 0 }}</div>
              <div>任务: {{ scope.row.statistics?.courierTasks || 0 }}</div>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column label="最后登录" width="160">
          <template #default="scope">
            <span v-if="scope.row.lastLogin">
              {{ formatDate(scope.row.lastLogin) }}
            </span>
            <span v-else class="text-gray">从未登录</span>
          </template>
        </el-table-column>
        
        <el-table-column label="注册时间" width="160">
          <template #default="scope">
            {{ formatDate(scope.row.createdAt) }}
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="scope">
            <el-button
              type="primary"
              size="small"
              @click="handleEdit(scope.row)"
              v-if="authStore.hasPermission('user.write')"
            >
              编辑
            </el-button>
            
            <el-button
              type="warning"
              size="small"
              @click="handleUnlock(scope.row)"
              v-if="scope.row.status === 'LOCKED' && authStore.hasPermission('user.write')"
            >
              解锁
            </el-button>
            
            <el-dropdown @command="(command) => handleUserAction(command, scope.row)">
              <el-button size="small">
                更多<el-icon class="el-icon--right"><arrow-down /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="resetPassword" v-if="authStore.hasPermission('user.write')">
                    重置密码
                  </el-dropdown-item>
                  <el-dropdown-item command="changeRole" v-if="authStore.hasPermission('user.role.manage')">
                    修改角色
                  </el-dropdown-item>
                  <el-dropdown-item command="viewDetails">
                    查看详情
                  </el-dropdown-item>
                  <el-dropdown-item command="delete" divided v-if="authStore.hasPermission('user.delete')">
                    删除用户
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.size"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <!-- 编辑用户对话框 -->
    <UserEditDialog
      v-model:visible="editDialogVisible"
      :user="currentUser"
      @saved="handleUserSaved"
    />

    <!-- 重置密码对话框 -->
    <ResetPasswordDialog
      v-model:visible="resetPasswordDialogVisible"
      :user="currentUser"
      @reset="handlePasswordReset"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { userApi } from '@/utils/api'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Download, ArrowDown } from '@element-plus/icons-vue'
import UserEditDialog from '@/components/UserEditDialog.vue'
import ResetPasswordDialog from '@/components/ResetPasswordDialog.vue'
import { formatDate } from '@/utils/date'
import type { User } from '@/types'

const authStore = useAuthStore()

// 响应式数据
const loading = ref(false)
const users = ref<User[]>([])
const selectedUsers = ref<User[]>([])
const editDialogVisible = ref(false)
const resetPasswordDialogVisible = ref(false)
const currentUser = ref<User | null>(null)

// 搜索表单
const searchForm = reactive({
  search: '',
  role: '',
  status: '',
  schoolCode: ''
})

// 分页信息
const pagination = reactive({
  page: 1,
  size: 20,
  total: 0
})

// 加载用户列表
const loadUsers = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      size: pagination.size,
      search: searchForm.search || undefined,
      role: searchForm.role || undefined,
      status: searchForm.status || undefined,
      schoolCode: searchForm.schoolCode || undefined
    }

    const response = await userApi.getUsers(params)
    if (response.data.code === 0) {
      users.value = response.data.data.items
      pagination.total = response.data.data.pagination.total
    }
  } catch (error) {
    ElMessage.error('加载用户列表失败')
  } finally {
    loading.value = false
  }
}

// 搜索处理
const handleSearch = () => {
  pagination.page = 1
  loadUsers()
}

// 重置搜索
const resetSearch = () => {
  Object.assign(searchForm, {
    search: '',
    role: '',
    status: '',
    schoolCode: ''
  })
  handleSearch()
}

// 分页处理
const handleSizeChange = (val: number) => {
  pagination.size = val
  pagination.page = 1
  loadUsers()
}

const handleCurrentChange = (val: number) => {
  pagination.page = val
  loadUsers()
}

// 选择变化处理
const handleSelectionChange = (val: User[]) => {
  selectedUsers.value = val
}

// 编辑用户
const handleEdit = (user: User) => {
  currentUser.value = user
  editDialogVisible.value = true
}

// 解锁用户
const handleUnlock = async (user: User) => {
  try {
    await ElMessageBox.confirm(`确定要解锁用户 ${user.username} 吗？`, '确认解锁', {
      type: 'warning'
    })
    
    await userApi.unlockUser(user.id)
    ElMessage.success('用户解锁成功')
    loadUsers()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('用户解锁失败')
    }
  }
}

// 用户操作处理
const handleUserAction = async (command: string, user: User) => {
  currentUser.value = user
  
  switch (command) {
    case 'resetPassword':
      resetPasswordDialogVisible.value = true
      break
    case 'changeRole':
      // TODO: 实现角色修改对话框
      ElMessage.info('角色修改功能开发中...')
      break
    case 'viewDetails':
      // TODO: 实现用户详情查看
      ElMessage.info('用户详情功能开发中...')
      break
    case 'delete':
      await handleDeleteUser(user)
      break
  }
}

// 删除用户
const handleDeleteUser = async (user: User) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除用户 ${user.username} 吗？此操作不可恢复。`,
      '确认删除',
      {
        type: 'error',
        confirmButtonText: '确定删除',
        confirmButtonClass: 'el-button--danger'
      }
    )
    
    await userApi.deleteUser(user.id)
    ElMessage.success('用户删除成功')
    loadUsers()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('用户删除失败')
    }
  }
}

// 用户保存处理
const handleUserSaved = () => {
  loadUsers()
}

// 密码重置处理
const handlePasswordReset = () => {
  loadUsers()
}

// 导出用户
const exportUsers = () => {
  ElMessage.info('用户导出功能开发中...')
}

// 获取角色标签类型
const getRoleTagType = (role: string) => {
  const types: Record<string, string> = {
    'super_admin': 'danger',
    'admin': 'warning',
    'courier': 'success',
    'user': 'info'
  }
  return types[role] || 'info'
}

// 获取角色文本
const getRoleText = (role: string) => {
  const texts: Record<string, string> = {
    'super_admin': '超级管理员',
    'admin': '管理员',
    'courier': '信使',
    'user': '普通用户'
  }
  return texts[role] || role
}

// 获取状态标签类型
const getStatusTagType = (status: string) => {
  const types: Record<string, string> = {
    'ACTIVE': 'success',
    'INACTIVE': 'info',
    'LOCKED': 'danger'
  }
  return types[status] || 'info'
}

// 获取状态文本
const getStatusText = (status: string) => {
  const texts: Record<string, string> = {
    'ACTIVE': '激活',
    'INACTIVE': '禁用',
    'LOCKED': '锁定'
  }
  return texts[status] || status
}

// 组件挂载时加载数据
onMounted(() => {
  loadUsers()
})
</script>

<style scoped>
.users-page {
  padding: 20px;
}

.page-header {
  margin-bottom: 20px;
}

.page-header h1 {
  margin: 0 0 8px 0;
  font-size: 24px;
  color: #333;
}

.page-header p {
  margin: 0;
  color: #666;
  font-size: 14px;
}

.filter-card {
  margin-bottom: 20px;
}

.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.table-actions {
  display: flex;
  gap: 8px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-details {
  flex: 1;
}

.username {
  font-weight: 500;
  color: #333;
  font-size: 14px;
}

.email {
  color: #666;
  font-size: 12px;
  margin-top: 2px;
}

.user-stats {
  font-size: 12px;
  color: #666;
}

.user-stats div {
  margin-bottom: 2px;
}

.text-gray {
  color: #999;
  font-style: italic;
}

.pagination-wrapper {
  margin-top: 20px;
  display: flex;
  justify-content: center;
}
</style>