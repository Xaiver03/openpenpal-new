<template>
  <div class="statistics-page">
    <div class="page-header">
      <h1>数据统计</h1>
      <p>系统运营数据分析和报表</p>
    </div>

    <!-- 统计概览 -->
    <el-row :gutter="20" class="overview-cards">
      <el-col :xs="24" :sm="12" :lg="6" v-for="stat in overviewStats" :key="stat.title">
        <el-card class="overview-card">
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
                <el-radio-button label="year">近一年</el-radio-button>
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

    <el-row :gutter="20" class="charts-section">
      <!-- 配送效率分析 -->
      <el-col :xs="24" :lg="16">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>配送效率分析</span>
              <el-select v-model="deliveryPeriod" size="small" style="width: 120px">
                <el-option label="本周" value="week" />
                <el-option label="本月" value="month" />
                <el-option label="本季度" value="quarter" />
              </el-select>
            </div>
          </template>
          <div class="chart-container">
            <v-chart class="chart" :option="deliveryEfficiencyOption" />
          </div>
        </el-card>
      </el-col>

      <!-- 学校分布 -->
      <el-col :xs="24" :lg="8">
        <el-card>
          <template #header>
            <span>学校用户分布</span>
          </template>
          <div class="school-stats">
            <div class="school-item" v-for="school in schoolStats" :key="school.name">
              <div class="school-info">
                <span class="school-name">{{ school.name }}</span>
                <span class="school-count">{{ school.count }}</span>
              </div>
              <div class="school-progress">
                <el-progress 
                  :percentage="school.percentage" 
                  :show-text="false" 
                  :stroke-width="6"
                />
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 详细数据表格 -->
    <el-card class="data-table-card">
      <template #header>
        <div class="card-header">
          <span>详细统计数据</span>
          <div class="table-controls">
            <el-date-picker
              v-model="dateRange"
              type="daterange"
              range-separator="至"
              start-placeholder="开始日期"
              end-placeholder="结束日期"
              size="small"
              style="width: 240px; margin-right: 10px"
            />
            <el-button type="primary" size="small" @click="refreshData">
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
            <el-button type="success" size="small" @click="exportData">
              <el-icon><Download /></el-icon>
              导出
            </el-button>
          </div>
        </div>
      </template>

      <el-table :data="detailStats" stripe>
        <el-table-column label="日期" prop="date" width="120" />
        <el-table-column label="新增用户" prop="newUsers" width="100" />
        <el-table-column label="活跃用户" prop="activeUsers" width="100" />
        <el-table-column label="新增信件" prop="newLetters" width="100" />
        <el-table-column label="成功投递" prop="deliveredLetters" width="100" />
        <el-table-column label="投递成功率" width="120">
          <template #default="scope">
            {{ ((scope.row.deliveredLetters / scope.row.newLetters) * 100).toFixed(1) }}%
          </template>
        </el-table-column>
        <el-table-column label="活跃信使" prop="activeCouriers" width="100" />
        <el-table-column label="平均配送时间" width="120">
          <template #default="scope">
            {{ scope.row.avgDeliveryTime }}小时
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, PieChart, BarChart } from 'echarts/charts'
import { 
  TitleComponent, 
  TooltipComponent, 
  LegendComponent, 
  GridComponent 
} from 'echarts/components'
import VChart from 'vue-echarts'
import { ElMessage } from 'element-plus'
import {
  User,
  Message,
  TruckFilled as Truck,
  DataAnalysis,
  Refresh,
  Download
} from '@element-plus/icons-vue'

// 注册ECharts组件
use([
  CanvasRenderer,
  LineChart,
  PieChart,
  BarChart,
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent
])

// 响应式数据
const userTrendPeriod = ref('month')
const deliveryPeriod = ref('week')
const dateRange = ref<[Date, Date] | null>(null)

// 概览统计数据
const overviewStats = ref([
  {
    title: '总用户数',
    value: '2,468',
    icon: 'User',
    color: '#409EFF'
  },
  {
    title: '总信件数',
    value: '15,287',
    icon: 'Message',
    color: '#67C23A'
  },
  {
    title: '活跃信使',
    value: '156',
    icon: 'Truck',
    color: '#E6A23C'
  },
  {
    title: '投递成功率',
    value: '94.2%',
    icon: 'DataAnalysis',
    color: '#F56C6C'
  }
])

// 用户增长趋势图表配置
const userTrendOption = ref({
  tooltip: {
    trigger: 'axis'
  },
  xAxis: {
    type: 'category',
    data: ['1月', '2月', '3月', '4月', '5月', '6月', '7月', '8月', '9月', '10月', '11月', '12月']
  },
  yAxis: {
    type: 'value'
  },
  series: [
    {
      name: '新增用户',
      data: [120, 132, 101, 134, 90, 230, 210, 320, 301, 254, 190, 330],
      type: 'line',
      smooth: true,
      itemStyle: {
        color: '#409EFF'
      }
    },
    {
      name: '活跃用户',
      data: [220, 182, 191, 234, 290, 330, 310, 420, 401, 354, 290, 430],
      type: 'line',
      smooth: true,
      itemStyle: {
        color: '#67C23A'
      }
    }
  ]
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
      { value: 3245, name: '已投递' },
      { value: 567, name: '运输中' },
      { value: 234, name: '已收取' },
      { value: 123, name: '草稿' },
      { value: 89, name: '失败' }
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

// 配送效率分析图表配置
const deliveryEfficiencyOption = ref({
  tooltip: {
    trigger: 'axis'
  },
  legend: {
    data: ['配送量', '成功率']
  },
  xAxis: {
    type: 'category',
    data: ['周一', '周二', '周三', '周四', '周五', '周六', '周日']
  },
  yAxis: [
    {
      type: 'value',
      name: '配送量',
      position: 'left'
    },
    {
      type: 'value',
      name: '成功率 (%)',
      position: 'right',
      min: 80,
      max: 100
    }
  ],
  series: [
    {
      name: '配送量',
      type: 'bar',
      data: [45, 52, 38, 67, 73, 89, 94],
      itemStyle: {
        color: '#409EFF'
      }
    },
    {
      name: '成功率',
      type: 'line',
      yAxisIndex: 1,
      data: [92, 94, 89, 96, 95, 97, 98],
      itemStyle: {
        color: '#67C23A'
      }
    }
  ]
})

// 学校统计数据
const schoolStats = ref([
  { name: '北京大学', count: 456, percentage: 85 },
  { name: '清华大学', count: 389, percentage: 72 },
  { name: '人民大学', count: 267, percentage: 50 },
  { name: '北京理工', count: 234, percentage: 43 },
  { name: '北京航空', count: 189, percentage: 35 },
  { name: '其他学校', count: 123, percentage: 23 }
])

// 详细统计数据
const detailStats = ref([
  {
    date: '2024-01-21',
    newUsers: 45,
    activeUsers: 234,
    newLetters: 123,
    deliveredLetters: 116,
    activeCouriers: 23,
    avgDeliveryTime: 2.5
  },
  {
    date: '2024-01-20',
    newUsers: 52,
    activeUsers: 267,
    newLetters: 156,
    deliveredLetters: 148,
    activeCouriers: 25,
    avgDeliveryTime: 2.8
  },
  {
    date: '2024-01-19',
    newUsers: 38,
    activeUsers: 198,
    newLetters: 89,
    deliveredLetters: 84,
    activeCouriers: 18,
    avgDeliveryTime: 3.2
  }
])

// 刷新数据
const refreshData = () => {
  ElMessage.success('数据刷新成功')
  // 这里可以调用API重新加载数据
}

// 导出数据
const exportData = () => {
  ElMessage.info('数据导出功能开发中...')
}

// 组件挂载时加载数据
onMounted(() => {
  // 这里可以调用API加载真实数据
  console.log('Statistics page mounted')
})
</script>

<style scoped>
.statistics-page {
  padding: 20px;
}

.page-header {
  margin-bottom: 24px;
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

.overview-cards {
  margin-bottom: 24px;
}

.overview-card {
  height: 120px;
}

.stat-content {
  display: flex;
  align-items: center;
  gap: 16px;
  height: 100%;
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

.school-stats {
  padding: 10px 0;
}

.school-item {
  margin-bottom: 16px;
}

.school-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.school-name {
  font-size: 14px;
  color: #333;
}

.school-count {
  font-size: 14px;
  font-weight: 600;
  color: #409EFF;
}

.school-progress {
  margin-bottom: 4px;
}

.data-table-card {
  margin-bottom: 24px;
}

.table-controls {
  display: flex;
  align-items: center;
  gap: 8px;
}
</style>