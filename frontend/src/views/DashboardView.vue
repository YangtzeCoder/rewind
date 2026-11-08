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
  <section class="dashboard"><div class="page-heading"><div><h1>概览</h1><p>查看你的备份空间。</p></div></div>
    <el-alert v-if="error" :title="error" type="error" show-icon :closable="false"><el-button link @click="load">重试</el-button></el-alert>
    <div v-loading="loading" class="stats-grid"><template v-if="data"><el-card><el-statistic title="项目" :value="data.project_count" /></el-card><el-card><el-statistic title="文件" :value="data.file_count" /></el-card><el-card><el-statistic title="备份版本" :value="data.version_count" /></el-card><el-card><el-statistic title="存储用量" :value="formatSize(data.total_size)" /></el-card></template></div>
    <el-card class="content-card"><div class="card-heading"><div><strong>管理你的项目</strong><p class="secondary-text">查看文件、浏览版本历史并下载备份。</p></div><el-button type="primary" @click="$router.push({ name: 'projects' })">查看项目</el-button></div></el-card>
  </section>
</template>
