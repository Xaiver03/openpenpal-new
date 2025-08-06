<template>
  <div class="not-found">
    <div class="not-found-content">
      <div class="not-found-image">
        <div class="error-code">404</div>
      </div>
      
      <div class="not-found-text">
        <h1>404</h1>
        <h2>页面未找到</h2>
        <p>抱歉，您访问的页面不存在或已被移除。</p>
      </div>
      
      <div class="not-found-actions">
        <el-button type="primary" @click="goHome">
          <el-icon><HomeFilled /></el-icon>
          返回首页
        </el-button>
        <el-button @click="goBack">
          <el-icon><Back /></el-icon>
          返回上页
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { HomeFilled, Back } from '@element-plus/icons-vue'

const router = useRouter()

const goHome = () => {
  router.push('/')
}

const goBack = () => {
  router.go(-1)
}
</script>

<style scoped>
.not-found {
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.not-found-content {
  text-align: center;
  max-width: 500px;
  padding: 40px;
}

.not-found-image {
  margin-bottom: 30px;
}

.not-found-image img {
  width: 200px;
  height: auto;
}

.not-found-text h1 {
  font-size: 80px;
  font-weight: 700;
  margin: 0;
  line-height: 1;
  background: linear-gradient(45deg, #fff, #f0f0f0);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.not-found-text h2 {
  font-size: 24px;
  font-weight: 600;
  margin: 10px 0;
}

.not-found-text p {
  font-size: 16px;
  margin: 20px 0;
  opacity: 0.9;
}

.not-found-actions {
  margin-top: 30px;
  display: flex;
  gap: 16px;
  justify-content: center;
}

@media (max-width: 768px) {
  .not-found-content {
    padding: 20px;
  }
  
  .not-found-text h1 {
    font-size: 60px;
  }
  
  .not-found-text h2 {
    font-size: 20px;
  }
  
  .not-found-actions {
    flex-direction: column;
    align-items: center;
  }
}
</style>