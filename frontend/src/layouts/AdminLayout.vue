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
  <div class="admin-shell">
    <aside class="sidebar">
      <div class="brand"><span class="brand-mark">R<span>↶</span></span><span>rewind<span class="brand-dot">.</span></span></div>
      <div class="nav-label">WORKSPACE</div>
      <nav aria-label="主导航">
        <RouterLink class="nav-item" :to="{ name: 'dashboard' }"><span class="nav-icon" aria-hidden="true">▦</span><span class="nav-text">概览</span></RouterLink>
        <RouterLink class="nav-item" :class="{ selected: route.name === 'project' }" :to="{ name: 'projects' }"><span class="nav-icon" aria-hidden="true">▤</span><span class="nav-text">项目</span></RouterLink>
        <RouterLink class="nav-item" :to="{ name: 'tokens' }"><span class="nav-icon" aria-hidden="true">⌘</span><span class="nav-text">API Token</span></RouterLink>
        <RouterLink class="nav-item" :to="{ name: 'settings' }"><span class="nav-icon" aria-hidden="true">⚙</span><span class="nav-text">账户设置</span></RouterLink>
      </nav>
      <div class="sidebar-bottom"><span class="status-dot"></span><span>自托管空间<small>YOUR PRIVATE WORKSPACE</small></span></div>
    </aside>
    <div class="admin-main">
      <header class="topbar"><div class="breadcrumb"><span>工作空间</span><span class="breadcrumb-divider">/</span><strong>{{ route.meta.title }}</strong></div><div class="account"><span class="avatar">{{ (auth.username || 'A').slice(0, 1).toUpperCase() }}</span><span class="account-name">{{ auth.username || 'admin' }}</span><el-button text :loading="loading" @click="logout">退出登录</el-button></div></header>
      <div v-if="error" class="logout-error" role="alert">{{ error }}</div>
      <main class="admin-content"><RouterView /></main>
      <footer class="admin-footer"><span>© Rewind</span><span>YOUR WORK, PRESERVED.</span></footer>
    </div>
  </div>
</template>
