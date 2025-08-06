<template>
  <el-dialog
    v-model="visible"
    title="编辑用户"
    width="600px"
    :before-close="handleClose"
  >
    <el-form ref="formRef" :model="form" :rules="rules" label-width="120px">
      <el-form-item label="用户名" prop="username">
        <el-input v-model="form.username" disabled />
      </el-form-item>
      <el-form-item label="邮箱" prop="email">
        <el-input v-model="form.email" />
      </el-form-item>
      <el-form-item label="角色" prop="role">
        <el-select v-model="form.role" placeholder="选择角色">
          <el-option label="普通用户" value="user" />
          <el-option label="信使" value="courier" />
          <el-option label="高级信使" value="senior_courier" />
          <el-option label="信使协调员" value="courier_coordinator" />
          <el-option label="学校管理员" value="school_admin" />
          <el-option label="平台管理员" value="platform_admin" />
          <el-option label="超级管理员" value="super_admin" />
        </el-select>
      </el-form-item>
      <el-form-item label="状态" prop="status">
        <el-select v-model="form.status" placeholder="选择状态">
          <el-option label="活跃" value="active" />
          <el-option label="锁定" value="locked" />
          <el-option label="停用" value="inactive" />
        </el-select>
      </el-form-item>
      <el-form-item label="学校代码" prop="schoolCode">
        <el-input v-model="form.schoolCode" placeholder="输入学校代码" />
      </el-form-item>
    </el-form>

    <template #footer>
      <span class="dialog-footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="loading">
          确定
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
  id: '',
  username: '',
  email: '',
  role: '',
  status: '',
  schoolCode: ''
})

const rules = {
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
  ],
  role: [
    { required: true, message: '请选择角色', trigger: 'change' }
  ],
  status: [
    { required: true, message: '请选择状态', trigger: 'change' }
  ]
}

watch(
  () => props.modelValue,
  (value) => {
    visible.value = value
    if (value && props.user) {
      Object.assign(form, props.user)
    }
  },
  { immediate: true }
)

watch(visible, (value) => {
  emit('update:modelValue', value)
})

const handleClose = () => {
  visible.value = false
  formRef.value?.resetFields()
}

const handleSubmit = async () => {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
    loading.value = true
    
    // TODO: 调用API更新用户
    await new Promise(resolve => setTimeout(resolve, 1000)) // 模拟API调用
    
    ElMessage.success('用户更新成功')
    emit('success')
    handleClose()
  } catch (error) {
    console.error('更新用户失败:', error)
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