<template>
  <div class="letters-page">
    <div class="page-header">
      <h1>信件管理</h1>
      <p>管理和监控系统中的所有信件</p>
    </div>

    <!-- 搜索和过滤 -->
    <el-card class="filter-card">
      <el-form :model="searchForm" inline>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="选择状态" clearable style="width: 150px">
            <el-option label="草稿" value="draft" />
            <el-option label="已生成" value="generated" />
            <el-option label="已收取" value="collected" />
            <el-option label="运输中" value="in_transit" />
            <el-option label="已投递" value="delivered" />
            <el-option label="失败" value="failed" />
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
        
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            format="YYYY-MM-DD"
            value-format="YYYY-MM-DD"
            style="width: 240px"
          />
        </el-form-item>
        
        <el-form-item label="紧急">
          <el-select v-model="searchForm.urgent" placeholder="是否紧急" clearable style="width: 100px">
            <el-option label="是" :value="true" />
            <el-option label="否" :value="false" />
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

    <!-- 信件列表 -->
    <el-card>
      <template #header>
        <div class="table-header">
          <span>信件列表</span>
          <div class="table-actions">
            <el-button type="success" size="small">
              <el-icon><Download /></el-icon>
              导出
            </el-button>
          </div>
        </div>
      </template>

      <el-table :data="letters" :loading="loading" stripe>
        <el-table-column label="信件ID" prop="id" width="150" />
        
        <el-table-column label="标题" prop="title" min-width="200" />
        
        <el-table-column label="发送者" width="150">
          <template #default="scope">
            <div>{{ scope.row.sender.username }}</div>
            <div class="text-gray text-sm">{{ scope.row.sender.schoolCode }}</div>
          </template>
        </el-table-column>
        
        <el-table-column label="状态" width="120">
          <template #default="scope">
            <el-tag :type="getStatusTagType(scope.row.status)" size="small">
              {{ getStatusText(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column label="紧急程度" width="100">
          <template #default="scope">
            <el-tag v-if="scope.row.urgent" type="danger" size="small">紧急</el-tag>
            <span v-else class="text-gray">普通</span>
          </template>
        </el-table-column>
        
        <el-table-column label="信使" width="120">
          <template #default="scope">
            <span v-if="scope.row.courier">{{ scope.row.courier.username }}</span>
            <span v-else class="text-gray">未分配</span>
          </template>
        </el-table-column>
        
        <el-table-column label="创建时间" width="160">
          <template #default="scope">
            {{ formatDate(scope.row.createdAt) }}
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="scope">
            <el-button type="primary" size="small" @click="handleViewDetails(scope.row)">
              详情
            </el-button>
            <el-button type="warning" size="small" @click="handleEditStatus(scope.row)">
              状态
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
import { letterApi } from '@/utils/api'
import { ElMessage } from 'element-plus'
import { Search, Download } from '@element-plus/icons-vue'
import { formatDate } from '@/utils/date'
import type { Letter } from '@/types'

// 响应式数据
const loading = ref(false)
const letters = ref<Letter[]>([])
const dateRange = ref<[string, string] | null>(null)

// 搜索表单
const searchForm = reactive({
  status: '',
  schoolCode: '',
  urgent: null as boolean | null
})

// 分页信息
const pagination = reactive({
  page: 1,
  size: 20,
  total: 0
})

// 加载信件列表
const loadLetters = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      size: pagination.size,
      status: searchForm.status || undefined,
      schoolCode: searchForm.schoolCode || undefined,
      urgent: searchForm.urgent,
      dateFrom: dateRange.value?.[0],
      dateTo: dateRange.value?.[1]
    }

    const response = await letterApi.getLetters(params)
    if (response.data.code === 0) {
      letters.value = response.data.data.items
      pagination.total = response.data.data.pagination.total
    }
  } catch (error) {
    ElMessage.error('加载信件列表失败')
  } finally {
    loading.value = false
  }
}

// 搜索处理
const handleSearch = () => {
  pagination.page = 1
  loadLetters()
}

// 重置搜索
const resetSearch = () => {
  Object.assign(searchForm, {
    status: '',
    schoolCode: '',
    urgent: null
  })
  dateRange.value = null
  handleSearch()
}

// 分页处理
const handleSizeChange = (val: number) => {
  pagination.size = val
  pagination.page = 1
  loadLetters()
}

const handleCurrentChange = (val: number) => {
  pagination.page = val
  loadLetters()
}

// 查看详情
const handleViewDetails = (letter: Letter) => {
  ElMessage.info('信件详情功能开发中...')
}

// 编辑状态
const handleEditStatus = (letter: Letter) => {
  ElMessage.info('状态编辑功能开发中...')
}

// 获取状态标签类型
const getStatusTagType = (status: string) => {
  const types: Record<string, string> = {
    'draft': 'info',
    'generated': 'primary',
    'collected': 'warning',
    'in_transit': 'success',
    'delivered': 'success',
    'failed': 'danger'
  }
  return types[status] || 'info'
}

// 获取状态文本
const getStatusText = (status: string) => {
  const texts: Record<string, string> = {
    'draft': '草稿',
    'generated': '已生成',
    'collected': '已收取',
    'in_transit': '运输中',
    'delivered': '已投递',
    'failed': '失败'
  }
  return texts[status] || status
}

// 组件挂载时加载数据
onMounted(() => {
  loadLetters()
})
</script>

<style scoped>
.letters-page {
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

.pagination-wrapper {
  margin-top: 20px;
  display: flex;
  justify-content: center;
}
</style>