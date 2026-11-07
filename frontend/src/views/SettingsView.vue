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
  error.value = ''; success.value = ''
  if (!current.value || !next.value) { error.value = '请输入当前密码和新密码。'; return }
  if (next.value !== confirm.value) { error.value = '两次输入的新密码不一致。'; return }
  if (new TextEncoder().encode(next.value).length > 72) { error.value = '新密码不能超过 72 字节。'; return }
  loading.value = true
  try { await api.changePassword(current.value, next.value); current.value = ''; next.value = ''; confirm.value = ''; success.value = '密码已更新。' } catch (e) { error.value = errorMessage(e) } finally { loading.value = false }
}
</script>
<template>
  <section class="dashboard"><div class="page-heading"><div><div class="eyebrow">ACCOUNT</div><h1>账户设置</h1><p>管理管理员账户的登录密码。</p></div></div>
    <form class="panel form-panel settings-form" @submit.prevent="submit"><h2>修改密码</h2><p v-if="error" class="form-error" role="alert">{{ error }}</p><p v-if="success" class="success-message" role="status">{{ success }}</p>
      <label for="current-password">当前密码</label><el-input id="current-password" v-model="current" type="password" show-password autocomplete="current-password" />
      <label for="new-password">新密码</label><el-input id="new-password" v-model="next" type="password" show-password autocomplete="new-password" />
      <label for="confirm-password">确认新密码</label><el-input id="confirm-password" v-model="confirm" type="password" show-password autocomplete="new-password" />
      <div class="form-actions"><el-button type="primary" native-type="submit" :loading="loading">更新密码</el-button></div>
    </form>
  </section>
</template>
