<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import type { BackupVersion, FileDetail } from '@/api'
import { useAuthStore } from '@/stores/auth'
import BackupPreviewDialog from '@/components/BackupPreviewDialog.vue'
import { errorMessage, formatDate, formatSize, saveFile } from './helpers'

const api = useAuthStore().api
const route = useRoute()
const file = ref<FileDetail>()
const versions = ref<BackupVersion[]>([])
const page = ref(1)
const loading = ref(false)
const error = ref('')
const downloading = ref('')
const preview = ref<InstanceType<typeof BackupPreviewDialog>>()
let requestNumber = 0
const projectId = () => String(route.params.id)

async function load() {
  const request = ++requestNumber
  const id = projectId()
  const fileId = String(route.params.fileId)
  loading.value = true; error.value = ''; file.value = undefined; versions.value = []
  try {
    const detail = await api.getFile(id, fileId)
    const history = await api.backupHistory({ project_id: id, key: detail.key, page: page.value, page_size: 20 })
    if (request === requestNumber) { file.value = detail; versions.value = history }
  } catch (e) {
    if (request === requestNumber) error.value = errorMessage(e)
  } finally {
    if (request === requestNumber) loading.value = false
  }
}
async function download(versionId: string) {
  downloading.value = versionId; error.value = ''
  try { const result = await api.downloadBackup(projectId(), versionId); saveFile(result.blob, result.filename) } catch (e) { error.value = errorMessage(e) } finally { downloading.value = '' }
}
function changePage(delta: number) { page.value += delta; void load() }
watch(() => [route.params.id, route.params.fileId], ([id, fileId]) => {
  if (id && fileId) { page.value = 1; void load() }
}, { immediate: true })
</script>

<template>
  <section class="dashboard">
    <el-button link type="primary" @click="$router.push({ name: 'project', params: { id: projectId() } })">← 返回项目详情</el-button>
    <div class="page-heading"><div><h1>文件详情</h1><p class="break-text">{{ file?.key }}</p></div></div>
    <el-alert v-if="error" :title="error" type="error" show-icon :closable="false"><el-button link @click="load">重试</el-button></el-alert>
    <el-card class="content-card"><template #header><div class="card-heading">版本历史</div></template>
      <el-table v-loading="loading" :data="versions"><el-table-column label="备注" min-width="180"><template #default="{ row }">{{ row.description || '暂无备注' }}</template></el-table-column><el-table-column label="备份时间" min-width="170"><template #default="{ row }">{{ formatDate(row.created_at) }}</template></el-table-column><el-table-column label="大小" min-width="100"><template #default="{ row }">{{ formatSize(row.size) }}</template></el-table-column><el-table-column label="操作" width="140"><template #default="{ row }"><el-button link type="primary" @click="preview?.open(projectId(), row.version_id)">预览</el-button><el-button link type="primary" :disabled="!!downloading" @click="download(row.version_id)">{{ downloading === row.version_id ? '下载中…' : '下载' }}</el-button></template></el-table-column><template #empty><el-empty description="没有更多版本。" /></template></el-table>
      <div class="pagination"><el-button :disabled="page === 1 || loading" @click="changePage(-1)">上一页</el-button><span>第 {{ page }} 页</span><el-button :disabled="versions.length < 20 || loading" @click="changePage(1)">下一页</el-button></div>
    </el-card>
    <BackupPreviewDialog ref="preview" />
  </section>
</template>
