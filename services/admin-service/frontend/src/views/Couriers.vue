<template>
  <div class="couriers-page">
    <div class="page-header">
      <h1>信使管理</h1>
      <p>管理系统中的所有信使和配送任务</p>
    </div>

    <!-- 搜索和过滤 -->
    <el-card class="filter-card">
      <el-form :model="searchForm" inline>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="选择状态" clearable style="width: 150px">
            <el-option label="待审核" value="pending" />
            <el-option label="已批准" value="approved" />
            <el-option label="活跃" value="active" />
            <el-option label="暂停" value="suspended" />
            <el-option label="禁用" value="banned" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="配送区域">
          <el-input
            v-model="searchForm.zone"
            placeholder="配送区域"
            clearable
            style="width: 150px"
          />
        </el-form-item>
        
        <el-form-item label="评分">
          <el-select v-model="searchForm.rating" placeholder="评分范围" clearable style="width: 120px">
            <el-option label="5星" value="5" />
            <el-option label="4星以上" value="4" />
            <el-option label="3星以上" value="3" />
            <el-option label="3星以下" value="low" />
          </el-select>
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

    <!-- 信使列表 -->
    <el-card>
      <template #header>
        <div class="table-header">
          <span>信使列表</span>
          <div class="table-actions">
            <el-button type="success" size="small">
              <el-icon><Download /></el-icon>
              导出
            </el-button>
          </div>
        </div>
      </template>

      <el-table :data="couriers" :loading="loading" stripe>
        <el-table-column label="信使信息" min-width="200">
          <template #default="scope">
            <div class="courier-info">
              <el-avatar size="small">
                {{ scope.row.user.username.charAt(0).toUpperCase() }}
              </el-avatar>
              <div class="courier-details">
                <div class="username">{{ scope.row.user.username }}</div>
                <div class="email">{{ scope.row.user.email }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column label="配送区域" prop="zone" width="120" />
        
        <el-table-column label="状态" width="100">
          <template #default="scope">
            <el-tag :type="getStatusTagType(scope.row.status)" size="small">
              {{ getStatusText(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column label="评分" width="120">
          <template #default="scope">
            <div class="rating">
              <el-rate
                :model-value="scope.row.rating"
                disabled
                show-score
                size="small"
              />
            </div>
          </template>
        </el-table-column>
        
        <el-table-column label="当前任务" width="100">
          <template #default="scope">
            <el-tag v-if="scope.row.currentTasks > 0" type="warning" size="small">
              {{ scope.row.currentTasks }}
            </el-tag>
            <span v-else class="text-gray">空闲</span>
          </template>
        </el-table-column>
        
        <el-table-column label="配送统计" width="200">
          <template #default="scope">
            <div class="delivery-stats">
              <div>总数: {{ scope.row.statistics.totalDeliveries }}</div>
              <div>成功率: {{ (scope.row.statistics.successRate * 100).toFixed(1) }}%</div>
              <div>平均时长: {{ formatDeliveryTime(scope.row.statistics.averageDeliveryTime) }}</div>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column label="最后活跃" width="160">
          <template #default="scope">
            {{ formatDate(scope.row.lastActive) }}
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="scope">
            <el-button
              v-if="scope.row.status === 'pending'"
              type="success"
              size="small"
              @click="handleApprove(scope.row)"
            >
              批准
            </el-button>
            <el-button
              v-if="scope.row.status === 'active'"
              type="warning"
              size="small"
              @click="handleSuspend(scope.row)"
            >
              暂停
            </el-button>
            <el-button
              v-if="scope.row.status === 'suspended'"
              type="primary"
              size="small"
              @click="handleActivate(scope.row)"
            >
              激活
            </el-button>
            <el-button
              type="primary"
              size="small"
              @click="handleViewDetails(scope.row)"
            >
              详情
            </el-button>
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
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { courierApi } from '@/utils/api'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Download } from '@element-plus/icons-vue'
import { formatDate } from '@/utils/date'
import type { Courier } from '@/types'

// 响应式数据
const loading = ref(false)
const couriers = ref<Courier[]>([])

// 搜索表单
const searchForm = reactive({
  status: '',
  zone: '',
  rating: ''
})

// 分页信息
const pagination = reactive({
  page: 1,
  size: 20,
  total: 0
})

// 加载信使列表
const loadCouriers = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      size: pagination.size,
      status: searchForm.status || undefined,
      zone: searchForm.zone || undefined,
      rating: searchForm.rating || undefined
    }

    const response = await courierApi.getCouriers(params)
    if (response.data.code === 0) {
      couriers.value = response.data.data.items
      pagination.total = response.data.data.pagination.total
    }
  } catch (error) {
    ElMessage.error('加载信使列表失败')
  } finally {
    loading.value = false
  }
}

// 搜索处理
const handleSearch = () => {
  pagination.page = 1
  loadCouriers()
}

// 重置搜索
const resetSearch = () => {
  Object.assign(searchForm, {
    status: '',
    zone: '',
    rating: ''
  })
  handleSearch()
}

// 分页处理
const handleSizeChange = (val: number) => {
  pagination.size = val
  pagination.page = 1
  loadCouriers()
}

const handleCurrentChange = (val: number) => {
  pagination.page = val
  loadCouriers()
}

// 批准信使
const handleApprove = async (courier: Courier) => {
  try {
    await ElMessageBox.confirm(`确定要批准信使 ${courier.user.username} 吗？`, '确认批准', {
      type: 'success'
    })
    
    await courierApi.updateCourierStatus(courier.id, 'approved')
    ElMessage.success('信使批准成功')
    loadCouriers()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('信使批准失败')
    }
  }
}

// 暂停信使
const handleSuspend = async (courier: Courier) => {
  try {
    await ElMessageBox.confirm(`确定要暂停信使 ${courier.user.username} 吗？`, '确认暂停', {
      type: 'warning'
    })
    
    await courierApi.updateCourierStatus(courier.id, 'suspended')
    ElMessage.success('信使暂停成功')
    loadCouriers()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('信使暂停失败')
    }
  }
}

// 激活信使
const handleActivate = async (courier: Courier) => {
  try {
    await courierApi.updateCourierStatus(courier.id, 'active')
    ElMessage.success('信使激活成功')
    loadCouriers()
  } catch (error) {
    ElMessage.error('信使激活失败')
  }
}

// 查看详情
const handleViewDetails = (courier: Courier) => {
  ElMessage.info('信使详情功能开发中...')
}

// 获取状态标签类型
const getStatusTagType = (status: string) => {
  const types: Record<string, string> = {
    'pending': 'warning',
    'approved': 'success',
    'active': 'success',
    'suspended': 'danger',
    'banned': 'danger'
  }
  return types[status] || 'info'
}

// 获取状态文本
const getStatusText = (status: string) => {
  const texts: Record<string, string> = {
    'pending': '待审核',
    'approved': '已批准',
    'active': '活跃',
    'suspended': '暂停',
    'banned': '禁用'
  }
  return texts[status] || status
}

// 格式化配送时间
const formatDeliveryTime = (hours: number) => {
  if (hours < 1) {
    return `${Math.round(hours * 60)}分钟`
  } else if (hours < 24) {
    return `${hours.toFixed(1)}小时`
  } else {
    return `${Math.round(hours / 24)}天`
  }
}

// 组件挂载时加载数据
onMounted(() => {
  loadCouriers()
})
</script>

<style scoped>
.couriers-page {
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

.courier-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.courier-details {
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

.rating {
  display: flex;
  align-items: center;
}

.delivery-stats {
  font-size: 12px;
  color: #666;
}

.delivery-stats div {
  margin-bottom: 2px;
}

.pagination-wrapper {
  margin-top: 20px;
  display: flex;
  justify-content: center;
}
</style>