<script setup lang="ts">
import { onMounted, ref } from 'vue'
import type { Dashboard } from '@/api'
import { useAuthStore } from '@/stores/auth'
import { errorMessage, formatSize } from './helpers'

const data = ref<Dashboard>()
const error = ref('')
const loading = ref(true)
async function load() {
  error.value = ''; loading.value = true
  try { data.value = await useAuthStore().api.dashboard() } catch (e) { error.value = errorMessage(e) } finally { loading.value = false }
}
onMounted(load)
</script>
<template>
  <section class="dashboard">
    <div class="page-heading"><div><div class="eyebrow">YOUR WORKSPACE</div><h1>概览</h1><p>查看你的备份空间。</p></div><div class="page-index">01 <span>/ OVERVIEW</span></div></div>
    <p v-if="error" class="form-error" role="alert">{{ error }} <button class="text-button" @click="load">重试</button></p>
    <p v-else-if="loading" class="muted">正在加载…</p>
    <div v-else-if="data" class="stats-grid">
      <div class="stat-card"><span>项目</span><strong>{{ data.project_count }}</strong><small>PROJECTS</small></div>
      <div class="stat-card"><span>文件</span><strong>{{ data.file_count }}</strong><small>FILES</small></div>
      <div class="stat-card"><span>备份版本</span><strong>{{ data.version_count }}</strong><small>VERSIONS</small></div>
      <div class="stat-card"><span>存储用量</span><strong>{{ formatSize(data.total_size) }}</strong><small>TOTAL SIZE</small></div>
    </div>
    <div class="dashboard-link"><div><h2>管理你的项目</h2><p>查看文件、浏览版本历史并下载备份。</p></div><RouterLink :to="{ name: 'projects' }">查看项目 →</RouterLink></div>
  </section>
</template>
