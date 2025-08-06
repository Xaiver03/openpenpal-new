<template>
  <div class="museum-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <h1>信件博物馆管理</h1>
      <p>管理信件展览、内容审核和敏感词过滤</p>
    </div>

    <!-- 功能标签页 -->
    <el-tabs v-model="activeTab" class="museum-tabs">
      <!-- 展览管理 -->
      <el-tab-pane label="展览管理" name="exhibitions">
        <div class="tab-content">
          <!-- 展览操作栏 -->
          <div class="actions-bar">
            <el-button type="primary" @click="showCreateExhibition = true">
              <el-icon><Plus /></el-icon>
              创建展览
            </el-button>
            <el-button @click="loadExhibitions">
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
          </div>

          <!-- 展览列表 -->
          <el-table
            :data="exhibitions"
            :loading="exhibitionsLoading"
            stripe
            class="exhibitions-table"
          >
            <el-table-column label="展览信息" min-width="200">
              <template #default="scope">
                <div class="exhibition-info">
                  <div class="title">{{ scope.row.title }}</div>
                  <div class="description">{{ scope.row.description }}</div>
                </div>
              </template>
            </el-table-column>
            
            <el-table-column label="状态" width="120">
              <template #default="scope">
                <el-tag :type="scope.row.isActive ? 'success' : 'info'">
                  {{ scope.row.isActive ? '活跃' : '停用' }}
                </el-tag>
              </template>
            </el-table-column>
            
            <el-table-column label="信件数量" width="100">
              <template #default="scope">
                <span>{{ scope.row.letterCount || 0 }}</span>
              </template>
            </el-table-column>
            
            <el-table-column label="创建时间" width="180">
              <template #default="scope">
                {{ formatDate(scope.row.createdAt) }}
              </template>
            </el-table-column>
            
            <el-table-column label="操作" width="200">
              <template #default="scope">
                <el-button size="small" @click="viewExhibition(scope.row)">查看</el-button>
                <el-button size="small" type="primary" @click="editExhibition(scope.row)">编辑</el-button>
                <el-button 
                  size="small" 
                  :type="scope.row.isActive ? 'warning' : 'success'"
                  @click="toggleExhibitionStatus(scope.row)"
                >
                  {{ scope.row.isActive ? '停用' : '激活' }}
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-tab-pane>

      <!-- 内容审核 -->
      <el-tab-pane label="内容审核" name="moderation">
        <div class="tab-content">
          <!-- 审核统计 -->
          <el-row :gutter="20" class="moderation-stats">
            <el-col :span="6">
              <el-card>
                <div class="stat-item">
                  <div class="stat-value">{{ moderationStats.pendingCount }}</div>
                  <div class="stat-label">待审核</div>
                </div>
              </el-card>
            </el-col>
            <el-col :span="6">
              <el-card>
                <div class="stat-item">
                  <div class="stat-value">{{ moderationStats.approvedCount }}</div>
                  <div class="stat-label">已通过</div>
                </div>
              </el-card>
            </el-col>
            <el-col :span="6">
              <el-card>
                <div class="stat-item">
                  <div class="stat-value">{{ moderationStats.rejectedCount }}</div>
                  <div class="stat-label">已拒绝</div>
                </div>
              </el-card>
            </el-col>
            <el-col :span="6">
              <el-card>
                <div class="stat-item">
                  <div class="stat-value">{{ moderationStats.reportedCount }}</div>
                  <div class="stat-label">被举报</div>
                </div>
              </el-card>
            </el-col>
          </el-row>

          <!-- 审核过滤器 -->
          <el-card class="filter-card">
            <el-form :model="moderationFilter" inline>
              <el-form-item label="审核状态">
                <el-select v-model="moderationFilter.status" placeholder="选择状态" clearable>
                  <el-option label="待审核" value="PENDING" />
                  <el-option label="已通过" value="APPROVED" />
                  <el-option label="已拒绝" value="REJECTED" />
                  <el-option label="被举报" value="REPORTED" />
                </el-select>
              </el-form-item>
              
              <el-form-item label="内容类型">
                <el-select v-model="moderationFilter.contentType" placeholder="选择类型" clearable>
                  <el-option label="信件内容" value="LETTER" />
                  <el-option label="用户评论" value="COMMENT" />
                  <el-option label="展览投稿" value="SUBMISSION" />
                </el-select>
              </el-form-item>
              
              <el-form-item>
                <el-button type="primary" @click="loadModerationTasks">搜索</el-button>
                <el-button @click="resetModerationFilter">重置</el-button>
              </el-form-item>
            </el-form>
          </el-card>

          <!-- 审核任务列表 -->
          <el-table
            :data="moderationTasks"
            :loading="moderationLoading"
            stripe
          >
            <el-table-column label="内容预览" min-width="300">
              <template #default="scope">
                <div class="content-preview">
                  <div class="content-text">{{ scope.row.contentPreview }}</div>
                  <div class="content-meta">
                    <el-tag size="small">{{ getContentTypeLabel(scope.row.contentType) }}</el-tag>
                    <span class="author">作者: {{ scope.row.authorName }}</span>
                  </div>
                </div>
              </template>
            </el-table-column>
            
            <el-table-column label="状态" width="100">
              <template #default="scope">
                <el-tag :type="getModerationStatusType(scope.row.status)">
                  {{ getModerationStatusLabel(scope.row.status) }}
                </el-tag>
              </template>
            </el-table-column>
            
            <el-table-column label="提交时间" width="180">
              <template #default="scope">
                {{ formatDate(scope.row.submittedAt) }}
              </template>
            </el-table-column>
            
            <el-table-column label="操作" width="200">
              <template #default="scope">
                <el-button 
                  size="small" 
                  type="success" 
                  @click="approveModerationTask(scope.row)"
                  :disabled="scope.row.status !== 'PENDING'"
                >
                  通过
                </el-button>
                <el-button 
                  size="small" 
                  type="danger" 
                  @click="rejectModerationTask(scope.row)"
                  :disabled="scope.row.status !== 'PENDING'"
                >
                  拒绝
                </el-button>
                <el-button size="small" @click="viewModerationDetail(scope.row)">详情</el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-tab-pane>

      <!-- 敏感词管理 -->
      <el-tab-pane label="敏感词管理" name="sensitive-words">
        <div class="tab-content">
          <!-- 敏感词操作栏 -->
          <div class="actions-bar">
            <el-button type="primary" @click="showAddSensitiveWord = true">
              <el-icon><Plus /></el-icon>
              添加敏感词
            </el-button>
            <el-button @click="loadSensitiveWords">
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
            <el-button type="warning" @click="batchDeleteSensitiveWords" :disabled="!selectedSensitiveWords.length">
              <el-icon><Delete /></el-icon>
              批量删除
            </el-button>
          </div>

          <!-- 敏感词列表 -->
          <el-table
            :data="sensitiveWords"
            :loading="sensitiveWordsLoading"
            stripe
            @selection-change="handleSensitiveWordSelection"
          >
            <el-table-column type="selection" width="55" />
            
            <el-table-column label="敏感词" prop="word" />
            
            <el-table-column label="类型" width="120">
              <template #default="scope">
                <el-tag :type="getSensitiveWordTypeColor(scope.row.type)">
                  {{ getSensitiveWordTypeLabel(scope.row.type) }}
                </el-tag>
              </template>
            </el-table-column>
            
            <el-table-column label="严重程度" width="120">
              <template #default="scope">
                <el-tag :type="getSeverityColor(scope.row.severity)">
                  {{ getSeverityLabel(scope.row.severity) }}
                </el-tag>
              </template>
            </el-table-column>
            
            <el-table-column label="状态" width="100">
              <template #default="scope">
                <el-tag :type="scope.row.isActive ? 'success' : 'info'">
                  {{ scope.row.isActive ? '启用' : '禁用' }}
                </el-tag>
              </template>
            </el-table-column>
            
            <el-table-column label="创建时间" width="180">
              <template #default="scope">
                {{ formatDate(scope.row.createdAt) }}
              </template>
            </el-table-column>
            
            <el-table-column label="操作" width="180">
              <template #default="scope">
                <el-button size="small" @click="editSensitiveWord(scope.row)">编辑</el-button>
                <el-button 
                  size="small" 
                  :type="scope.row.isActive ? 'warning' : 'success'"
                  @click="toggleSensitiveWordStatus(scope.row)"
                >
                  {{ scope.row.isActive ? '禁用' : '启用' }}
                </el-button>
                <el-button size="small" type="danger" @click="deleteSensitiveWord(scope.row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-tab-pane>
    </el-tabs>

    <!-- 创建展览对话框 -->
    <el-dialog
      v-model="showCreateExhibition"
      title="创建展览"
      width="600px"
    >
      <el-form :model="newExhibition" :rules="exhibitionRules" ref="exhibitionFormRef" label-width="100px">
        <el-form-item label="展览标题" prop="title">
          <el-input v-model="newExhibition.title" placeholder="请输入展览标题" />
        </el-form-item>
        
        <el-form-item label="展览描述" prop="description">
          <el-input 
            v-model="newExhibition.description" 
            type="textarea" 
            :rows="4"
            placeholder="请输入展览描述"
          />
        </el-form-item>
        
        <el-form-item label="主题标签">
          <el-input v-model="newExhibition.theme" placeholder="请输入主题标签" />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="showCreateExhibition = false">取消</el-button>
        <el-button type="primary" @click="createExhibition" :loading="creating">创建</el-button>
      </template>
    </el-dialog>

    <!-- 添加敏感词对话框 -->
    <el-dialog
      v-model="showAddSensitiveWord"
      title="添加敏感词"
      width="500px"
    >
      <el-form :model="newSensitiveWord" :rules="sensitiveWordRules" ref="sensitiveWordFormRef" label-width="100px">
        <el-form-item label="敏感词" prop="word">
          <el-input v-model="newSensitiveWord.word" placeholder="请输入敏感词" />
        </el-form-item>
        
        <el-form-item label="类型" prop="type">
          <el-select v-model="newSensitiveWord.type" placeholder="选择类型">
            <el-option label="违法违规" value="ILLEGAL" />
            <el-option label="色情内容" value="PORNOGRAPHIC" />
            <el-option label="暴力血腥" value="VIOLENT" />
            <el-option label="涉政敏感" value="POLITICAL" />
            <el-option label="其他不当" value="OTHER" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="严重程度" prop="severity">
          <el-select v-model="newSensitiveWord.severity" placeholder="选择严重程度">
            <el-option label="低" value="LOW" />
            <el-option label="中" value="MEDIUM" />
            <el-option label="高" value="HIGH" />
            <el-option label="严重" value="CRITICAL" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="处理动作" prop="action">
          <el-select v-model="newSensitiveWord.action" placeholder="选择处理动作">
            <el-option label="标记审核" value="REVIEW" />
            <el-option label="自动屏蔽" value="BLOCK" />
            <el-option label="替换字符" value="REPLACE" />
          </el-select>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="showAddSensitiveWord = false">取消</el-button>
        <el-button type="primary" @click="addSensitiveWord" :loading="adding">添加</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { formatDate } from '@/utils/date'
import { Plus, Refresh, Delete } from '@element-plus/icons-vue'
import { museumApi } from '@/utils/api'

// 活跃标签页
const activeTab = ref('exhibitions')

// 展览管理相关
const exhibitions = ref<any[]>([])
const exhibitionsLoading = ref(false)
const showCreateExhibition = ref(false)
const creating = ref(false)
const newExhibition = ref({
  title: '',
  description: '',
  theme: ''
})

const exhibitionRules = {
  title: [{ required: true, message: '请输入展览标题', trigger: 'blur' }],
  description: [{ required: true, message: '请输入展览描述', trigger: 'blur' }]
}

// 内容审核相关
const moderationTasks = ref<any[]>([])
const moderationLoading = ref(false)
const moderationStats = ref({
  pendingCount: 0,
  approvedCount: 0,
  rejectedCount: 0,
  reportedCount: 0
})
const moderationFilter = ref({
  status: '',
  contentType: ''
})

// 敏感词管理相关
const sensitiveWords = ref<any[]>([])
const sensitiveWordsLoading = ref(false)
const selectedSensitiveWords = ref<any[]>([])
const showAddSensitiveWord = ref(false)
const adding = ref(false)
const newSensitiveWord = ref({
  word: '',
  type: '',
  severity: '',
  action: ''
})

const sensitiveWordRules = {
  word: [{ required: true, message: '请输入敏感词', trigger: 'blur' }],
  type: [{ required: true, message: '请选择类型', trigger: 'change' }],
  severity: [{ required: true, message: '请选择严重程度', trigger: 'change' }],
  action: [{ required: true, message: '请选择处理动作', trigger: 'change' }]
}

// 加载展览列表
const loadExhibitions = async () => {
  exhibitionsLoading.value = true
  try {
    const response = await museumApi.getExhibitions()
    if (response.data.code === 0) {
      exhibitions.value = response.data.data.items || []
    }
  } catch (error) {
    ElMessage.error('加载展览列表失败')
  } finally {
    exhibitionsLoading.value = false
  }
}

// 创建展览
const createExhibition = async () => {
  creating.value = true
  try {
    const response = await museumApi.createExhibition(newExhibition.value)
    if (response.data.code === 0) {
      ElMessage.success('展览创建成功')
      showCreateExhibition.value = false
      newExhibition.value = { title: '', description: '', theme: '' }
      loadExhibitions()
    }
  } catch (error) {
    ElMessage.error('创建展览失败')
  } finally {
    creating.value = false
  }
}

// 切换展览状态
const toggleExhibitionStatus = async (exhibition: any) => {
  try {
    const action = exhibition.isActive ? '停用' : '激活'
    await ElMessageBox.confirm(`确定要${action}展览"${exhibition.title}"吗？`, '确认操作')
    
    const response = await museumApi.updateExhibitionStatus(exhibition.id, {
      isActive: !exhibition.isActive
    })
    
    if (response.data.code === 0) {
      ElMessage.success(`展览${action}成功`)
      loadExhibitions()
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('操作失败')
    }
  }
}

// 加载审核统计
const loadModerationStats = async () => {
  try {
    const response = await museumApi.getModerationStats()
    if (response.data.code === 0) {
      moderationStats.value = response.data.data
    }
  } catch (error) {
    console.error('加载审核统计失败:', error)
  }
}

// 加载审核任务
const loadModerationTasks = async () => {
  moderationLoading.value = true
  try {
    const params = {
      ...moderationFilter.value,
      page: 1,
      limit: 50
    }
    const response = await museumApi.getModerationTasks(params)
    if (response.data.code === 0) {
      moderationTasks.value = response.data.data.items || []
    }
  } catch (error) {
    ElMessage.error('加载审核任务失败')
  } finally {
    moderationLoading.value = false
  }
}

// 审核通过
const approveModerationTask = async (task: any) => {
  try {
    const response = await museumApi.approveModerationTask(task.id)
    if (response.data.code === 0) {
      ElMessage.success('审核通过')
      loadModerationTasks()
      loadModerationStats()
    }
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

// 审核拒绝
const rejectModerationTask = async (task: any) => {
  try {
    const reason = await ElMessageBox.prompt('请输入拒绝原因:', '拒绝审核')
    const response = await museumApi.rejectModerationTask(task.id, {
      reason: reason.value
    })
    if (response.data.code === 0) {
      ElMessage.success('审核拒绝')
      loadModerationTasks()
      loadModerationStats()
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('操作失败')
    }
  }
}

// 重置审核过滤器
const resetModerationFilter = () => {
  moderationFilter.value = { status: '', contentType: '' }
  loadModerationTasks()
}

// 加载敏感词列表
const loadSensitiveWords = async () => {
  sensitiveWordsLoading.value = true
  try {
    const response = await museumApi.getSensitiveWords()
    if (response.data.code === 0) {
      sensitiveWords.value = response.data.data.items || []
    }
  } catch (error) {
    ElMessage.error('加载敏感词列表失败')
  } finally {
    sensitiveWordsLoading.value = false
  }
}

// 添加敏感词
const addSensitiveWord = async () => {
  adding.value = true
  try {
    const response = await museumApi.addSensitiveWord(newSensitiveWord.value)
    if (response.data.code === 0) {
      ElMessage.success('敏感词添加成功')
      showAddSensitiveWord.value = false
      newSensitiveWord.value = { word: '', type: '', severity: '', action: '' }
      loadSensitiveWords()
    }
  } catch (error) {
    ElMessage.error('添加敏感词失败')
  } finally {
    adding.value = false
  }
}

// 删除敏感词
const deleteSensitiveWord = async (word: any) => {
  try {
    await ElMessageBox.confirm(`确定要删除敏感词"${word.word}"吗？`, '确认删除')
    const response = await museumApi.deleteSensitiveWord(word.id)
    if (response.data.code === 0) {
      ElMessage.success('删除成功')
      loadSensitiveWords()
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

// 切换敏感词状态
const toggleSensitiveWordStatus = async (word: any) => {
  try {
    const response = await museumApi.updateSensitiveWordStatus(word.id, {
      isActive: !word.isActive
    })
    if (response.data.code === 0) {
      ElMessage.success('状态更新成功')
      loadSensitiveWords()
    }
  } catch (error) {
    ElMessage.error('状态更新失败')
  }
}

// 敏感词选择变化
const handleSensitiveWordSelection = (selection: any[]) => {
  selectedSensitiveWords.value = selection
}

// 批量删除敏感词
const batchDeleteSensitiveWords = async () => {
  try {
    await ElMessageBox.confirm(`确定要删除选中的 ${selectedSensitiveWords.value.length} 个敏感词吗？`, '确认批量删除')
    const ids = selectedSensitiveWords.value.map(word => word.id)
    const response = await museumApi.batchDeleteSensitiveWords({ ids })
    if (response.data.code === 0) {
      ElMessage.success('批量删除成功')
      loadSensitiveWords()
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('批量删除失败')
    }
  }
}

// 辅助函数
const getContentTypeLabel = (type: string) => {
  const labels: Record<string, string> = {
    LETTER: '信件内容',
    COMMENT: '用户评论',
    SUBMISSION: '展览投稿'
  }
  return labels[type] || type
}

const getModerationStatusType = (status: string) => {
  const types: Record<string, string> = {
    PENDING: 'warning',
    APPROVED: 'success',
    REJECTED: 'danger',
    REPORTED: 'info'
  }
  return types[status] || 'info'
}

const getModerationStatusLabel = (status: string) => {
  const labels: Record<string, string> = {
    PENDING: '待审核',
    APPROVED: '已通过',
    REJECTED: '已拒绝',
    REPORTED: '被举报'
  }
  return labels[status] || status
}

const getSensitiveWordTypeColor = (type: string) => {
  const colors: Record<string, string> = {
    ILLEGAL: 'danger',
    PORNOGRAPHIC: 'danger',
    VIOLENT: 'warning',
    POLITICAL: 'info',
    OTHER: ''
  }
  return colors[type] || ''
}

const getSensitiveWordTypeLabel = (type: string) => {
  const labels: Record<string, string> = {
    ILLEGAL: '违法违规',
    PORNOGRAPHIC: '色情内容',
    VIOLENT: '暴力血腥',
    POLITICAL: '涉政敏感',
    OTHER: '其他不当'
  }
  return labels[type] || type
}

const getSeverityColor = (severity: string) => {
  const colors: Record<string, string> = {
    LOW: 'info',
    MEDIUM: 'warning',
    HIGH: 'danger',
    CRITICAL: 'danger'
  }
  return colors[severity] || ''
}

const getSeverityLabel = (severity: string) => {
  const labels: Record<string, string> = {
    LOW: '低',
    MEDIUM: '中',
    HIGH: '高',
    CRITICAL: '严重'
  }
  return labels[severity] || severity
}

// 生命周期
onMounted(() => {
  loadExhibitions()
  loadModerationStats()
  loadModerationTasks()
  loadSensitiveWords()
})

// 其他函数占位符
const viewExhibition = (exhibition: any) => {
  ElMessage.info('查看展览功能开发中')
}

const editExhibition = (exhibition: any) => {
  ElMessage.info('编辑展览功能开发中')
}

const viewModerationDetail = (task: any) => {
  ElMessage.info('查看审核详情功能开发中')
}

const editSensitiveWord = (word: any) => {
  ElMessage.info('编辑敏感词功能开发中')
}
</script>

<style scoped>
.museum-page {
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

.museum-tabs {
  background: white;
}

.tab-content {
  padding: 20px;
}

.actions-bar {
  margin-bottom: 20px;
  display: flex;
  gap: 12px;
}

.filter-card {
  margin-bottom: 20px;
}

.exhibition-info .title {
  font-weight: 600;
  color: #333;
  margin-bottom: 4px;
}

.exhibition-info .description {
  color: #666;
  font-size: 12px;
}

.moderation-stats {
  margin-bottom: 24px;
}

.stat-item {
  text-align: center;
}

.stat-value {
  font-size: 24px;
  font-weight: 600;
  color: #409EFF;
  line-height: 1;
}

.stat-label {
  font-size: 14px;
  color: #666;
  margin-top: 8px;
}

.content-preview .content-text {
  font-size: 14px;
  color: #333;
  margin-bottom: 8px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.content-preview .content-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 12px;
  color: #666;
}

.exhibitions-table,
.moderation-table,
.sensitive-words-table {
  margin-top: 16px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-details .username {
  font-weight: 600;
  color: #333;
  font-size: 14px;
}

.user-details .email {
  color: #666;
  font-size: 12px;
}
</style>