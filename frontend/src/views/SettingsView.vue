<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { errorMessage } from './helpers'

const api = useAuthStore().api
const current = ref('')
const next = ref('')
const confirm = ref('')
const loading = ref(false)
const error = ref('')
const success = ref('')
async function submit() {
  error.value = ''
  success.value = ''
  if (!current.value || !next.value) {
    error.value = '请输入当前密码和新密码。'
    return
  }
  if (next.value !== confirm.value) {
    error.value = '两次输入的新密码不一致。'
    return
  }
  if (new TextEncoder().encode(next.value).length > 72) {
    error.value = '新密码不能超过 72 字节。'
    return
  }
  loading.value = true
  try {
    await api.changePassword(current.value, next.value)
    current.value = ''
    next.value = ''
    confirm.value = ''
    success.value = '密码已更新。'
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    loading.value = false
  }
}
</script>
<template>
  <section class="dashboard">
    <div class="page-heading">
      <div>
        <h1>账户设置</h1>
        <p>管理管理员账户的登录密码。</p>
      </div>
    </div>
    <el-card class="content-card settings-card">
      <template #header>修改密码</template>
      <el-alert v-if="error" :title="error" type="error" show-icon :closable="false" />
      <el-alert v-if="success" :title="success" type="success" show-icon :closable="false" />
      <el-form label-position="top" @submit.prevent="submit">
        <el-form-item label="当前密码">
          <el-input
            v-model="current"
            type="password"
            show-password
            autocomplete="current-password"
          />
        </el-form-item>
        <el-form-item label="新密码">
          <el-input v-model="next" type="password" show-password autocomplete="new-password" />
        </el-form-item>
        <el-form-item label="确认新密码">
          <el-input v-model="confirm" type="password" show-password autocomplete="new-password" />
        </el-form-item>
        <el-button type="primary" native-type="submit" :loading="loading">更新密码</el-button>
      </el-form>
    </el-card>
  </section>
</template>
