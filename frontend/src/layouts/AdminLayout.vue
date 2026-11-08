<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()
function expired() { void router.replace({ name: 'login', query: { expired: '1' } }) }
onMounted(() => window.addEventListener('rewind:unauthorized', expired))
onUnmounted(() => window.removeEventListener('rewind:unauthorized', expired))
const loading = ref(false)
const error = ref('')
async function logout() {
  loading.value = true
  error.value = ''
  try {
    await auth.logout()
    await router.replace({ name: 'login' })
  } catch {
    error.value = '退出失败，请检查网络后重试。'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <el-container class="admin-shell">
    <el-aside class="sidebar"><div class="brand"><span class="brand-mark">R</span><span class="brand-name">rewind<span class="brand-dot">.</span></span></div>
      <el-menu router :default-active="route.name === 'project' ? '/projects' : route.path" class="navigation" aria-label="主导航"><el-menu-item index="/dashboard"><span class="nav-icon">▦</span><span class="nav-text">概览</span></el-menu-item><el-menu-item index="/projects"><span class="nav-icon">▤</span><span class="nav-text">项目</span></el-menu-item><el-menu-item index="/tokens"><span class="nav-icon">⌘</span><span class="nav-text">API Token</span></el-menu-item><el-menu-item index="/settings"><span class="nav-icon">⚙</span><span class="nav-text">账户设置</span></el-menu-item></el-menu>
    </el-aside>
    <el-container class="admin-main"><el-header class="topbar"><el-breadcrumb separator="/"><el-breadcrumb-item>工作空间</el-breadcrumb-item><el-breadcrumb-item>{{ route.meta.title }}</el-breadcrumb-item></el-breadcrumb><div class="account"><el-avatar :size="32">{{ (auth.username || 'A').slice(0, 1).toUpperCase() }}</el-avatar><span class="account-name">{{ auth.username || 'admin' }}</span><el-button text :loading="loading" @click="logout">退出登录</el-button></div></el-header>
      <el-alert v-if="error" :title="error" type="error" show-icon :closable="false" /><el-main class="admin-content"><RouterView /></el-main><el-footer class="admin-footer">© Rewind</el-footer>
    </el-container>
  </el-container>
</template>
