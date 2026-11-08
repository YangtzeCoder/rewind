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
    <el-button link type="primary" @click="$router.push({ name: 'projects' })">← 返回项目</el-button>
    <div class="page-heading"><div><h1>{{ project?.name || '项目详情' }}</h1><p>{{ project?.description || '查看项目中的文件和备份版本。' }}</p></div></div>
    <el-alert v-if="error" :title="error" type="error" show-icon :closable="false"><el-button link @click="load">重试</el-button></el-alert>
    <el-card class="content-card"><template #header><div class="card-heading"><span>文件</span><el-form class="search-form" @submit.prevent="filter"><el-input v-model="search" placeholder="按文件 key 筛选" aria-label="按文件 key 筛选" clearable /><el-button native-type="submit">搜索</el-button></el-form></div></template>
      <el-table v-loading="loading" :data="files"><el-table-column label="文件 key" min-width="220"><template #default="{ row }"><el-button link type="primary" @click="history(row)">{{ row.key }}</el-button></template></el-table-column><el-table-column label="版本" min-width="120"><template #default="{ row }">{{ row.version_count }} 个版本</template></el-table-column><el-table-column label="最新大小" min-width="110"><template #default="{ row }">{{ row.latest_version ? formatSize(row.latest_version.size) : '—' }}</template></el-table-column><el-table-column label="更新时间" min-width="170"><template #default="{ row }">{{ formatDate(row.updated_at) }}</template></el-table-column><el-table-column label="操作" width="180"><template #default="{ row }"><el-button link type="primary" @click="history(row)">版本历史</el-button><el-button v-if="row.latest_version" link type="primary" :disabled="!!downloading" @click="download(row.latest_version.version_id)">{{ downloading === row.latest_version.version_id ? '下载中…' : '下载最新' }}</el-button></template></el-table-column><template #empty><el-empty description="没有找到文件。" /></template></el-table>
      <div class="pagination"><el-button :disabled="page === 1 || loading" @click="changePage(-1)">上一页</el-button><span>第 {{ page }} 页</span><el-button :disabled="files.length < 20 || loading" @click="changePage(1)">下一页</el-button></div>
    </el-card>
    <el-card v-if="selected" id="history" class="content-card"><template #header><div class="card-heading"><div>版本历史 <span class="secondary-text break-text">{{ selected.key }}</span></div><el-button link @click="selected = undefined">关闭</el-button></div></template><el-table v-loading="historyLoading" :data="versions"><el-table-column label="备注" min-width="180"><template #default="{ row }">{{ row.description || '暂无备注' }}</template></el-table-column><el-table-column label="备份时间" min-width="170"><template #default="{ row }">{{ formatDate(row.created_at) }}</template></el-table-column><el-table-column label="大小" min-width="100"><template #default="{ row }">{{ formatSize(row.size) }}</template></el-table-column><el-table-column label="操作" width="90"><template #default="{ row }"><el-button link type="primary" :disabled="!!downloading" @click="download(row.version_id)">{{ downloading === row.version_id ? '下载中…' : '下载' }}</el-button></template></el-table-column><template #empty><el-empty description="没有更多版本。" /></template></el-table><div class="pagination"><el-button :disabled="historyPage === 1 || historyLoading" @click="changeHistoryPage(-1)">上一页</el-button><span>第 {{ historyPage }} 页</span><el-button :disabled="versions.length < 20 || historyLoading" @click="changeHistoryPage(1)">下一页</el-button></div></el-card>
  </section>
</template>
