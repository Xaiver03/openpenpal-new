<template>
  <div class="dashboard">
    <div class="dashboard-header">
      <h1>系统概览</h1>
      <p>欢迎回来，{{ authStore.user?.username }}！</p>
    </div>

    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stats-cards">
      <el-col :xs="24" :sm="12" :md="6" v-for="stat in statsCards" :key="stat.title">
        <el-card class="stat-card" :body-style="{ padding: '20px' }">
          <div class="stat-content">
            <div class="stat-icon" :style="{ backgroundColor: stat.color }">
              <el-icon :size="24">
                <component :is="stat.icon" />
              </el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stat.value }}</div>
              <div class="stat-title">{{ stat.title }}</div>
            </div>
          </div>
          <div class="stat-change" :class="stat.change >= 0 ? 'positive' : 'negative'">
            <el-icon>
              <TrendCharts v-if="stat.change >= 0" />
              <Bottom v-else />
            </el-icon>
            <span>{{ Math.abs(stat.change) }}%</span>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 图表区域 -->
    <el-row :gutter="20" class="charts-section">
      <!-- 用户增长趋势 -->
      <el-col :xs="24" :lg="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>用户增长趋势</span>
              <el-radio-group v-model="userTrendPeriod" size="small">
                <el-radio-button label="week">近一周</el-radio-button>
                <el-radio-button label="month">近一月</el-radio-button>
              </el-radio-group>
            </div>
          </template>
          <div class="chart-container">
            <v-chart class="chart" :option="userTrendOption" />
          </div>
        </el-card>
      </el-col>

      <!-- 信件状态分布 -->
      <el-col :xs="24" :lg="12">
        <el-card>
          <template #header>
            <span>信件状态分布</span>
          </template>
          <div class="chart-container">
            <v-chart class="chart" :option="letterStatusOption" />
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 最近活动 -->
    <el-row :gutter="20" class="activities-section">
      <el-col :xs="24" :lg="16">
        <el-card>
          <template #header>
            <span>最近活动</span>
          </template>
          <el-timeline>
            <el-timeline-item
              v-for="activity in recentActivities"
              :key="activity.id"
              :timestamp="activity.timestamp"
              :type="activity.type as 'primary' | 'success' | 'warning' | 'danger' | 'info'"
            >
              <div class="activity-content">
                <strong>{{ activity.user }}</strong>
                <span>{{ activity.action }}</span>
                <em>{{ activity.target }}</em>
              </div>
            </el-timeline-item>
          </el-timeline>
        </el-card>
      </el-col>

      <!-- 系统状态 -->
      <el-col :xs="24" :lg="8">
        <el-card>
          <template #header>
            <span>系统状态</span>
          </template>
          <div class="system-status">
            <div class="status-item" v-for="status in systemStatus" :key="status.name">
              <div class="status-info">
                <span class="status-name">{{ status.name }}</span>
                <el-tag :type="status.status === 'healthy' ? 'success' : 'danger'" size="small">
                  {{ status.status === 'healthy' ? '正常' : '异常' }}
                </el-tag>
              </div>
              <div class="status-value">{{ status.value }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, PieChart } from 'echarts/charts'
import { 
  TitleComponent, 
  TooltipComponent, 
  LegendComponent, 
  GridComponent 
} from 'echarts/components'
import VChart from 'vue-echarts'
import {
  User,
  Message,
  Van,
  DataAnalysis,
  TrendCharts,
  Bottom
} from '@element-plus/icons-vue'

// 注册ECharts组件
use([
  CanvasRenderer,
  LineChart,
  PieChart,
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent
])

const authStore = useAuthStore()

// 用户趋势图表时间段
const userTrendPeriod = ref('week')

// 统计卡片数据
const statsCards = ref([
  {
    title: '总用户数',
    value: '1,248',
    icon: 'User',
    color: '#409EFF',
    change: 12.5
  },
  {
    title: '活跃信件',
    value: '856',
    icon: 'Message',
    color: '#67C23A',
    change: 8.2
  },
  {
    title: '在线信使',
    value: '32',
    icon: 'Van',
    color: '#E6A23C',
    change: -2.1
  },
  {
    title: '今日投递',
    value: '156',
    icon: 'DataAnalysis',
    color: '#F56C6C',
    change: 15.3
  }
])

// 用户增长趋势图表配置
const userTrendOption = ref({
  tooltip: {
    trigger: 'axis'
  },
  xAxis: {
    type: 'category',
    data: ['周一', '周二', '周三', '周四', '周五', '周六', '周日']
  },
  yAxis: {
    type: 'value'
  },
  series: [{
    data: [120, 132, 101, 134, 90, 230, 210],
    type: 'line',
    smooth: true,
    itemStyle: {
      color: '#409EFF'
    }
  }]
})

// 信件状态分布图表配置
const letterStatusOption = ref({
  tooltip: {
    trigger: 'item'
  },
  legend: {
    bottom: '0%',
    left: 'center'
  },
  series: [{
    type: 'pie',
    radius: ['40%', '70%'],
    center: ['50%', '45%'],
    data: [
      { value: 897, name: '已投递' },
      { value: 89, name: '运输中' },
      { value: 156, name: '已收取' },
      { value: 45, name: '草稿' },
      { value: 38, name: '失败' }
    ],
    emphasis: {
      itemStyle: {
        shadowBlur: 10,
        shadowOffsetX: 0,
        shadowColor: 'rgba(0, 0, 0, 0.5)'
      }
    }
  }]
})

// 最近活动数据
const recentActivities = ref([
  {
    id: 1,
    user: '管理员',
    action: '更新了用户',
    target: 'alice@example.com',
    timestamp: '2024-01-21 15:30:00',
    type: 'primary'
  },
  {
    id: 2,
    user: 'system',
    action: '处理了异常信件',
    target: 'OP1K2L3M4N5O',
    timestamp: '2024-01-21 14:45:00',
    type: 'warning'
  },
  {
    id: 3,
    user: '信使管理员',
    action: '激活了信使',
    target: 'courier_456',
    timestamp: '2024-01-21 13:20:00',
    type: 'success'
  }
])

// 系统状态数据
const systemStatus = ref([
  {
    name: '数据库',
    status: 'healthy',
    value: '正常连接'
  },
  {
    name: 'Redis缓存',
    status: 'healthy',
    value: '正常连接'
  },
  {
    name: 'CPU使用率',
    status: 'healthy',
    value: '45%'
  },
  {
    name: '内存使用率',
    status: 'healthy',
    value: '67%'
  }
])

// 组件挂载时加载数据
onMounted(() => {
  // 这里可以调用API加载真实数据
  console.log('Dashboard mounted')
})
</script>

<style scoped>
.dashboard {
  padding: 20px;
}

.dashboard-header {
  margin-bottom: 24px;
}

.dashboard-header h1 {
  margin: 0 0 8px 0;
  font-size: 24px;
  color: #333;
}

.dashboard-header p {
  margin: 0;
  color: #666;
  font-size: 14px;
}

.stats-cards {
  margin-bottom: 24px;
}

.stat-card {
  height: 120px;
}

.stat-content {
  display: flex;
  align-items: center;
  gap: 16px;
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 24px;
  font-weight: 600;
  color: #333;
  line-height: 1;
}

.stat-title {
  font-size: 14px;
  color: #666;
  margin-top: 4px;
}

.stat-change {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  margin-top: 8px;
}

.stat-change.positive {
  color: #67C23A;
}

.stat-change.negative {
  color: #F56C6C;
}

.charts-section {
  margin-bottom: 24px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.chart-container {
  height: 300px;
}

.chart {
  height: 100%;
}

.activities-section {
  margin-bottom: 24px;
}

.activity-content {
  display: flex;
  gap: 8px;
  align-items: center;
}

.activity-content strong {
  color: #409EFF;
}

.activity-content em {
  color: #666;
  font-style: normal;
  background: #f5f5f5;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 12px;
}

.system-status {
  space-y: 16px;
}

.status-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid #f0f0f0;
}

.status-item:last-child {
  border-bottom: none;
}

.status-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.status-name {
  font-size: 14px;
  color: #333;
}

.status-value {
  font-size: 12px;
  color: #666;
}
</style>