<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import type { BackupVersion, Project, ProjectFile } from '@/api'
import { useAuthStore } from '@/stores/auth'
import { errorMessage, formatDate, formatSize, saveFile } from './helpers'

const api = useAuthStore().api
const route = useRoute()
const project = ref<Project>()
const files = ref<ProjectFile[]>([])
const page = ref(1)
const search = ref('')
const activeSearch = ref('')
const loading = ref(false)
const error = ref('')
const selected = ref<ProjectFile>()
const versions = ref<BackupVersion[]>([])
const historyPage = ref(1)
const historyLoading = ref(false)
const downloading = ref('')
const id = () => String(route.params.id)
async function loadFiles() {
  loading.value = true; error.value = ''; files.value = []
  try { files.value = await api.listFiles(id(), { page: page.value, page_size: 20, key: activeSearch.value }) } catch (e) { error.value = errorMessage(e) } finally { loading.value = false }
}
async function load() {
  project.value = undefined; selected.value = undefined; error.value = ''
  try { project.value = await api.getProject(id()); await loadFiles() } catch (e) { error.value = errorMessage(e) }
}
function filter() { page.value = 1; activeSearch.value = search.value; void loadFiles() }
async function history(file: ProjectFile, reset = true) {
  selected.value = file
  if (reset) historyPage.value = 1
  historyLoading.value = true; error.value = ''; versions.value = []
  try { versions.value = await api.backupHistory({ project_id: id(), key: file.key, page: historyPage.value, page_size: 20 }) } catch (e) { error.value = errorMessage(e) } finally { historyLoading.value = false }
}
async function download(versionId: string) {
  downloading.value = versionId; error.value = ''
  try { const file = await api.downloadBackup(id(), versionId); saveFile(file.blob, file.filename) } catch (e) { error.value = errorMessage(e) } finally { downloading.value = '' }
}
function changePage(delta: number) { page.value += delta; void loadFiles() }
function changeHistoryPage(delta: number) { historyPage.value += delta; if (selected.value) void history(selected.value, false) }
onMounted(load)
watch(() => route.params.id, (value) => { if (value) { page.value = 1; search.value = ''; activeSearch.value = ''; void load() } })
</script>
<template>
  <section class="dashboard">
    <RouterLink class="back-link" :to="{ name: 'projects' }">← 返回项目</RouterLink>
    <div class="page-heading"><div><div class="eyebrow">PROJECT</div><h1>{{ project?.name || '项目详情' }}</h1><p>{{ project?.description || '查看项目中的文件和备份版本。' }}</p><small v-if="project" class="muted">{{ project.id }}</small></div></div>
    <p v-if="error" class="form-error" role="alert">{{ error }} <button class="text-button" @click="load">重试</button></p>
    <div class="panel list-panel">
      <div class="panel-title"><h2>文件</h2><form class="search-form" @submit.prevent="filter"><el-input v-model="search" placeholder="按文件 key 筛选" aria-label="按文件 key 筛选" clearable /><el-button native-type="submit">搜索</el-button></form></div>
      <p v-if="loading" class="empty-text">正在加载…</p><p v-else-if="!files.length" class="empty-text">没有找到文件。</p>
      <div v-for="file in files" v-else :key="file.id" class="list-row"><div class="row-main"><button class="item-title link-button" @click="history(file)">{{ file.key }}</button><p>{{ file.version_count }} 个版本<span v-if="file.latest_version"> · 最新 {{ formatSize(file.latest_version.size) }}</span></p><small>更新于 {{ formatDate(file.updated_at) }}</small></div><div class="row-actions"><button class="text-button" @click="history(file)">版本历史</button><button v-if="file.latest_version" class="text-button" :disabled="!!downloading" @click="download(file.latest_version.version_id)">{{ downloading === file.latest_version.version_id ? '下载中…' : '下载最新' }}</button></div></div>
      <div class="pagination"><el-button :disabled="page === 1 || loading" @click="changePage(-1)">上一页</el-button><span>第 {{ page }} 页</span><el-button :disabled="files.length < 20 || loading" @click="changePage(1)">下一页</el-button></div>
    </div>
    <div v-if="selected" class="panel list-panel" id="history"><div class="panel-title"><div><h2>版本历史</h2><p class="muted key-text">{{ selected.key }}</p></div><button class="text-button" @click="selected = undefined">关闭</button></div>
      <p v-if="historyLoading" class="empty-text">正在加载…</p><p v-else-if="!versions.length" class="empty-text">没有更多版本。</p>
      <div v-for="version in versions" v-else :key="version.version_id" class="list-row"><div class="row-main"><strong class="item-title">{{ version.version_id }}</strong><p>{{ version.description || '暂无描述' }}</p><small>{{ formatDate(version.created_at) }} · {{ formatSize(version.size) }}</small></div><button class="text-button" :disabled="!!downloading" @click="download(version.version_id)">{{ downloading === version.version_id ? '下载中…' : '下载' }}</button></div>
      <div class="pagination"><el-button :disabled="historyPage === 1 || historyLoading" @click="changeHistoryPage(-1)">上一页</el-button><span>第 {{ historyPage }} 页</span><el-button :disabled="versions.length < 20 || historyLoading" @click="changeHistoryPage(1)">下一页</el-button></div>
    </div>
  </section>
</template>
