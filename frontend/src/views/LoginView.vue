<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ApiError } from '@/api'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()
const username = ref('')
const password = ref('')
const loading = ref(false)
const error = ref(route.query.expired ? '登录已过期，请重新登录。' : route.query.unavailable ? '无法连接服务，请检查网络后重试。' : '')

async function submit() {
  if (!username.value.trim() || !password.value || loading.value) return
  error.value = ''
  loading.value = true
  try {
    await auth.login(username.value, password.value)
    await router.replace({ name: 'dashboard' })
  } catch (cause) {
    error.value = cause instanceof ApiError && cause.code === 'INVALID_CREDENTIALS'
      ? '用户名或密码不正确，请重试。'
      : '登录失败，请检查服务连接后重试。'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <main class="login-page">
    <div class="login-shell">
      <div class="brand login-brand"><span class="brand-mark">R<span>↶</span></span><span>rewind<span class="brand-dot">.</span></span></div>
      <div class="login-card">
        <div class="eyebrow"><span class="status-dot"></span> REWIND CONSOLE</div>
        <h1>欢迎回来</h1>
        <p class="login-intro">登录管理后台，管理你的备份与版本。</p>
        <el-form label-position="top" @submit.prevent="submit">
          <el-alert v-if="error" :title="error" type="error" show-icon :closable="false" />
          <el-form-item label="用户名">
          <el-input id="username" v-model="username" size="large" autocomplete="username" placeholder="请输入用户名" />
          </el-form-item><el-form-item label="密码">
          <el-input id="password" v-model="password" size="large" type="password" show-password autocomplete="current-password" placeholder="请输入密码" />
          </el-form-item><el-button class="login-button" type="primary" size="large" native-type="submit" :loading="loading" :disabled="!username.trim() || !password">登录 <span aria-hidden="true">→</span></el-button>
        </el-form>
      </div>
      <p class="login-footer"><span>为你的工作留一份安心。</span><span>SELF-HOSTED · PRIVATE BY DESIGN</span></p>
    </div>
    <div class="login-decoration" aria-hidden="true"><span>01 / YOUR WORK, PRESERVED</span><div class="orbit orbit-one"></div><div class="orbit orbit-two"></div><div class="orbit orbit-three"></div><div class="decoration-copy">Every version<br><em>has a story.</em></div><small>保留每一次改变，随时回到你需要的时刻。</small></div>
  </main>
</template>
