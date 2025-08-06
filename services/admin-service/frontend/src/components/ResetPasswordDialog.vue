<template>
  <el-dialog
    v-model="visible"
    title="重置密码"
    width="500px"
    :before-close="handleClose"
  >
    <el-form ref="formRef" :model="form" :rules="rules" label-width="120px">
      <el-form-item label="用户名">
        <el-input :value="user?.username" disabled />
      </el-form-item>
      <el-form-item label="新密码" prop="password">
        <el-input
          v-model="form.password"
          type="password"
          placeholder="请输入新密码"
          show-password
        />
      </el-form-item>
      <el-form-item label="确认密码" prop="confirmPassword">
        <el-input
          v-model="form.confirmPassword"
          type="password"
          placeholder="请再次输入密码"
          show-password
        />
      </el-form-item>
    </el-form>

    <template #footer>
      <span class="dialog-footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="loading">
          确定重置
        </el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { ElMessage, type FormInstance } from 'element-plus'

interface Props {
  modelValue: boolean
  user?: any
}

interface Emits {
  (e: 'update:modelValue', value: boolean): void
  (e: 'success'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const formRef = ref<FormInstance>()
const loading = ref(false)
const visible = ref(false)

const form = reactive({
  password: '',
  confirmPassword: ''
})

const validatePassword = (rule: any, value: string, callback: any) => {
  if (value.length < 6) {
    callback(new Error('密码长度至少6位'))
  } else {
    if (form.confirmPassword !== '') {
      formRef.value?.validateField('confirmPassword')
    }
    callback()
  }
}

const validateConfirmPassword = (rule: any, value: string, callback: any) => {
  if (value !== form.password) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const rules = {
  password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { validator: validatePassword, trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' }
  ]
}

watch(
  () => props.modelValue,
  (value) => {
    visible.value = value
  },
  { immediate: true }
)

watch(visible, (value) => {
  emit('update:modelValue', value)
})

const handleClose = () => {
  visible.value = false
  formRef.value?.resetFields()
  form.password = ''
  form.confirmPassword = ''
}

const handleSubmit = async () => {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
    loading.value = true
    
    // TODO: 调用API重置密码
    await new Promise(resolve => setTimeout(resolve, 1000)) // 模拟API调用
    
    ElMessage.success('密码重置成功')
    emit('success')
    handleClose()
  } catch (error) {
    console.error('重置密码失败:', error)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
</style>