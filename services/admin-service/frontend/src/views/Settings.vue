<template>
  <div class="settings-page">
    <div class="page-header">
      <h1>系统设置</h1>
      <p>配置系统参数和管理选项</p>
    </div>

    <el-row :gutter="20">
      <!-- 设置菜单 -->
      <el-col :xs="24" :md="6">
        <el-card class="settings-menu">
          <el-menu
            :default-active="activeTab"
            @select="handleMenuSelect"
            class="settings-nav"
          >
            <el-menu-item index="general">
              <el-icon><Setting /></el-icon>
              <span>基本设置</span>
            </el-menu-item>
            <el-menu-item index="security">
              <el-icon><Lock /></el-icon>
              <span>安全设置</span>
            </el-menu-item>
            <el-menu-item index="notification">
              <el-icon><Bell /></el-icon>
              <span>通知设置</span>
            </el-menu-item>
            <el-menu-item index="mail">
              <el-icon><Message /></el-icon>
              <span>邮件设置</span>
            </el-menu-item>
            <el-menu-item index="system">
              <el-icon><Monitor /></el-icon>
              <span>系统信息</span>
            </el-menu-item>
          </el-menu>
        </el-card>
      </el-col>

      <!-- 设置内容 -->
      <el-col :xs="24" :md="18">
        <!-- 基本设置 -->
        <el-card v-show="activeTab === 'general'" class="settings-content">
          <template #header>
            <span>基本设置</span>
          </template>
          
          <el-form :model="generalSettings" label-width="150px">
            <el-form-item label="系统名称">
              <el-input v-model="generalSettings.systemName" placeholder="请输入系统名称" />
            </el-form-item>
            
            <el-form-item label="系统描述">
              <el-input
                v-model="generalSettings.systemDescription"
                type="textarea"
                :rows="3"
                placeholder="请输入系统描述"
              />
            </el-form-item>
            
            <el-form-item label="维护模式">
              <el-switch
                v-model="generalSettings.maintenanceMode"
                active-text="开启"
                inactive-text="关闭"
              />
              <div class="form-help">开启后普通用户无法访问系统</div>
            </el-form-item>
            
            <el-form-item label="用户注册">
              <el-switch
                v-model="generalSettings.allowRegistration"
                active-text="允许"
                inactive-text="禁止"
              />
            </el-form-item>
            
            <el-form-item label="匿名信件">
              <el-switch
                v-model="generalSettings.allowAnonymousLetters"
                active-text="允许"
                inactive-text="禁止"
              />
            </el-form-item>
            
            <el-form-item label="每日信件限制">
              <el-input-number
                v-model="generalSettings.maxLettersPerDay"
                :min="1"
                :max="100"
                controls-position="right"
              />
              <span class="form-unit">封/天</span>
            </el-form-item>
            
            <el-form-item>
              <el-button type="primary" @click="saveGeneralSettings" :loading="saving">
                保存设置
              </el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <!-- 安全设置 -->
        <el-card v-show="activeTab === 'security'" class="settings-content">
          <template #header>
            <span>安全设置</span>
          </template>
          
          <el-form :model="securitySettings" label-width="150px">
            <el-form-item label="密码最小长度">
              <el-input-number
                v-model="securitySettings.minPasswordLength"
                :min="6"
                :max="32"
                controls-position="right"
              />
              <span class="form-unit">字符</span>
            </el-form-item>
            
            <el-form-item label="密码复杂度">
              <el-switch
                v-model="securitySettings.requireComplexPassword"
                active-text="强制"
                inactive-text="不强制"
              />
              <div class="form-help">包含大小写字母、数字和特殊字符</div>
            </el-form-item>
            
            <el-form-item label="最大登录尝试">
              <el-input-number
                v-model="securitySettings.maxLoginAttempts"
                :min="3"
                :max="10"
                controls-position="right"
              />
              <span class="form-unit">次</span>
            </el-form-item>
            
            <el-form-item label="账户锁定时间">
              <el-input-number
                v-model="securitySettings.lockoutDuration"
                :min="5"
                :max="1440"
                controls-position="right"
              />
              <span class="form-unit">分钟</span>
            </el-form-item>
            
            <el-form-item label="会话超时">
              <el-input-number
                v-model="securitySettings.sessionTimeout"
                :min="10"
                :max="480"
                controls-position="right"
              />
              <span class="form-unit">分钟</span>
            </el-form-item>
            
            <el-form-item>
              <el-button type="primary" @click="saveSecuritySettings" :loading="saving">
                保存设置
              </el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <!-- 通知设置 -->
        <el-card v-show="activeTab === 'notification'" class="settings-content">
          <template #header>
            <span>通知设置</span>
          </template>
          
          <el-form :model="notificationSettings" label-width="150px">
            <el-form-item label="邮件通知">
              <el-switch
                v-model="notificationSettings.emailEnabled"
                active-text="启用"
                inactive-text="禁用"
              />
            </el-form-item>
            
            <el-form-item label="系统异常通知">
              <el-switch
                v-model="notificationSettings.systemErrorNotification"
                active-text="启用"
                inactive-text="禁用"
              />
            </el-form-item>
            
            <el-form-item label="用户注册通知">
              <el-switch
                v-model="notificationSettings.userRegistrationNotification"
                active-text="启用"
                inactive-text="禁用"
              />
            </el-form-item>
            
            <el-form-item label="信件异常通知">
              <el-switch
                v-model="notificationSettings.letterErrorNotification"
                active-text="启用"
                inactive-text="禁用"
              />
            </el-form-item>
            
            <el-form-item>
              <el-button type="primary" @click="saveNotificationSettings" :loading="saving">
                保存设置
              </el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <!-- 邮件设置 -->
        <el-card v-show="activeTab === 'mail'" class="settings-content">
          <template #header>
            <span>邮件设置</span>
          </template>
          
          <el-form :model="mailSettings" label-width="150px">
            <el-form-item label="SMTP服务器">
              <el-input v-model="mailSettings.smtpHost" placeholder="smtp.example.com" />
            </el-form-item>
            
            <el-form-item label="SMTP端口">
              <el-input-number
                v-model="mailSettings.smtpPort"
                :min="1"
                :max="65535"
                controls-position="right"
              />
            </el-form-item>
            
            <el-form-item label="用户名">
              <el-input v-model="mailSettings.smtpUsername" placeholder="username@example.com" />
            </el-form-item>
            
            <el-form-item label="密码">
              <el-input
                v-model="mailSettings.smtpPassword"
                type="password"
                placeholder="SMTP密码"
                show-password
              />
            </el-form-item>
            
            <el-form-item label="SSL/TLS">
              <el-switch
                v-model="mailSettings.enableSsl"
                active-text="启用"
                inactive-text="禁用"
              />
            </el-form-item>
            
            <el-form-item label="发件人名称">
              <el-input v-model="mailSettings.fromName" placeholder="OpenPenPal System" />
            </el-form-item>
            
            <el-form-item>
              <el-button type="primary" @click="saveMailSettings" :loading="saving">
                保存设置
              </el-button>
              <el-button @click="testMail" :loading="testing">
                测试邮件
              </el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <!-- 系统信息 -->
        <el-card v-show="activeTab === 'system'" class="settings-content">
          <template #header>
            <span>系统信息</span>
          </template>
          
          <el-descriptions :column="2" border>
            <el-descriptions-item label="系统版本">v1.0.0</el-descriptions-item>
            <el-descriptions-item label="数据库版本">PostgreSQL 14.0</el-descriptions-item>
            <el-descriptions-item label="Java版本">OpenJDK 17.0.2</el-descriptions-item>
            <el-descriptions-item label="Spring Boot版本">3.2.1</el-descriptions-item>
            <el-descriptions-item label="服务器时间">{{ currentTime }}</el-descriptions-item>
            <el-descriptions-item label="系统启动时间">2024-01-21 09:00:00</el-descriptions-item>
            <el-descriptions-item label="运行时长">{{ uptime }}</el-descriptions-item>
            <el-descriptions-item label="内存使用">2.1GB / 4.0GB</el-descriptions-item>
          </el-descriptions>
          
          <div class="system-actions">
            <el-button type="warning" @click="clearCache">
              <el-icon><Delete /></el-icon>
              清理缓存
            </el-button>
            <el-button type="info" @click="backupDatabase">
              <el-icon><DocumentCopy /></el-icon>
              备份数据库
            </el-button>
            <el-button type="success" @click="checkUpdates">
              <el-icon><Refresh /></el-icon>
              检查更新
            </el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import {
  Setting,
  Lock,
  Bell,
  Message,
  Monitor,
  Delete,
  DocumentCopy,
  Refresh
} from '@element-plus/icons-vue'
import { formatDate } from '@/utils/date'

// 响应式数据
const activeTab = ref('general')
const saving = ref(false)
const testing = ref(false)
const currentTime = ref(formatDate(new Date()))
const uptime = ref('1天 15小时 32分钟')

// 设置数据
const generalSettings = reactive({
  systemName: 'OpenPenPal',
  systemDescription: '校园信件投递管理系统',
  maintenanceMode: false,
  allowRegistration: true,
  allowAnonymousLetters: true,
  maxLettersPerDay: 10
})

const securitySettings = reactive({
  minPasswordLength: 8,
  requireComplexPassword: true,
  maxLoginAttempts: 5,
  lockoutDuration: 30,
  sessionTimeout: 120
})

const notificationSettings = reactive({
  emailEnabled: true,
  systemErrorNotification: true,
  userRegistrationNotification: false,
  letterErrorNotification: true
})

const mailSettings = reactive({
  smtpHost: '',
  smtpPort: 587,
  smtpUsername: '',
  smtpPassword: '',
  enableSsl: true,
  fromName: 'OpenPenPal System'
})

// 更新时间定时器
let timeTimer: number | null = null

// 菜单选择处理
const handleMenuSelect = (index: string) => {
  activeTab.value = index
}

// 保存基本设置
const saveGeneralSettings = async () => {
  saving.value = true
  try {
    // 这里调用API保存设置
    await new Promise(resolve => setTimeout(resolve, 1000))
    ElMessage.success('基本设置保存成功')
  } catch (error) {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

// 保存安全设置
const saveSecuritySettings = async () => {
  saving.value = true
  try {
    // 这里调用API保存设置
    await new Promise(resolve => setTimeout(resolve, 1000))
    ElMessage.success('安全设置保存成功')
  } catch (error) {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

// 保存通知设置
const saveNotificationSettings = async () => {
  saving.value = true
  try {
    // 这里调用API保存设置
    await new Promise(resolve => setTimeout(resolve, 1000))
    ElMessage.success('通知设置保存成功')
  } catch (error) {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

// 保存邮件设置
const saveMailSettings = async () => {
  saving.value = true
  try {
    // 这里调用API保存设置
    await new Promise(resolve => setTimeout(resolve, 1000))
    ElMessage.success('邮件设置保存成功')
  } catch (error) {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

// 测试邮件
const testMail = async () => {
  testing.value = true
  try {
    // 这里调用API测试邮件
    await new Promise(resolve => setTimeout(resolve, 2000))
    ElMessage.success('测试邮件发送成功')
  } catch (error) {
    ElMessage.error('测试邮件发送失败')
  } finally {
    testing.value = false
  }
}

// 清理缓存
const clearCache = () => {
  ElMessage.success('缓存清理成功')
}

// 备份数据库
const backupDatabase = () => {
  ElMessage.info('数据库备份功能开发中...')
}

// 检查更新
const checkUpdates = () => {
  ElMessage.success('系统已是最新版本')
}

// 更新当前时间
const updateTime = () => {
  currentTime.value = formatDate(new Date())
}

// 组件挂载时设置定时器
onMounted(() => {
  timeTimer = window.setInterval(updateTime, 1000)
})

// 组件卸载时清理定时器
onUnmounted(() => {
  if (timeTimer) {
    clearInterval(timeTimer)
  }
})
</script>

<style scoped>
.settings-page {
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

.settings-menu {
  margin-bottom: 20px;
}

.settings-nav {
  border: none;
}

.settings-nav .el-menu-item {
  border-radius: 6px;
  margin-bottom: 4px;
}

.settings-content {
  min-height: 400px;
}

.form-help {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
}

.form-unit {
  margin-left: 8px;
  color: #666;
  font-size: 14px;
}

.system-actions {
  margin-top: 24px;
  display: flex;
  gap: 12px;
}

@media (max-width: 768px) {
  .system-actions {
    flex-direction: column;
  }
  
  .system-actions .el-button {
    margin-left: 0;
    margin-bottom: 8px;
  }
}
</style>