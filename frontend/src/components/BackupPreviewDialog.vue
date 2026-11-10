<script setup lang="ts">
import { onBeforeUnmount, ref } from 'vue'
import type { DownloadResult } from '@/api'
import { useAuthStore } from '@/stores/auth'
import { errorMessage, saveFile } from '@/views/helpers'

type PreviewKind = 'text' | 'image' | 'audio' | 'video' | 'unsupported'

const api = useAuthStore().api
const visible = ref(false)
const loading = ref(false)
const error = ref('')
const kind = ref<PreviewKind>('unsupported')
const text = ref('')
const objectUrl = ref('')
const result = ref<DownloadResult>()
const projectId = ref('')
const versionId = ref('')
let requestNumber = 0

const extensionTypes: Record<string, string> = {
  txt: 'text/plain',
  log: 'text/plain',
  md: 'text/markdown',
  markdown: 'text/markdown',
  csv: 'text/csv',
  tsv: 'text/tab-separated-values',
  html: 'text/html',
  htm: 'text/html',
  css: 'text/css',
  js: 'text/javascript',
  mjs: 'text/javascript',
  ts: 'text/plain',
  tsx: 'text/plain',
  jsx: 'text/plain',
  json: 'application/json',
  jsonl: 'application/json',
  xml: 'application/xml',
  xhtml: 'application/xhtml+xml',
  yaml: 'text/plain',
  yml: 'text/plain',
  toml: 'text/plain',
  ini: 'text/plain',
  conf: 'text/plain',
  sh: 'text/plain',
  py: 'text/plain',
  go: 'text/plain',
  rs: 'text/plain',
  java: 'text/plain',
  c: 'text/plain',
  h: 'text/plain',
  cpp: 'text/plain',
  hpp: 'text/plain',
  sql: 'text/plain',
  vue: 'text/plain',
  svg: 'image/svg+xml',
  jpg: 'image/jpeg',
  jpeg: 'image/jpeg',
  png: 'image/png',
  gif: 'image/gif',
  webp: 'image/webp',
  avif: 'image/avif',
  apng: 'image/apng',
  bmp: 'image/bmp',
  ico: 'image/x-icon',
  mp3: 'audio/mpeg',
  m4a: 'audio/mp4',
  aac: 'audio/aac',
  oga: 'audio/ogg',
  opus: 'audio/ogg',
  wav: 'audio/wav',
  weba: 'audio/webm',
  flac: 'audio/flac',
  mp4: 'video/mp4',
  webm: 'video/webm',
  ogv: 'video/ogg',
  mov: 'video/quicktime',
  mpeg: 'video/mpeg',
  mpg: 'video/mpeg',
}
const images = new Set([
  'image/jpeg',
  'image/png',
  'image/gif',
  'image/webp',
  'image/avif',
  'image/apng',
  'image/bmp',
  'image/x-icon',
  'image/vnd.microsoft.icon',
  'image/svg+xml',
])
const audio = new Set([
  'audio/mpeg',
  'audio/mp4',
  'audio/aac',
  'audio/ogg',
  'audio/opus',
  'audio/wav',
  'audio/webm',
  'audio/flac',
])
const video = new Set(['video/mp4', 'video/webm', 'video/ogg', 'video/quicktime', 'video/mpeg'])
const textApplications = new Set([
  'application/json',
  'application/xml',
  'application/javascript',
  'application/xhtml+xml',
  'application/x-sh',
  'application/x-yaml',
  'application/yaml',
  'application/toml',
  'application/sql',
])

function previewType(file: DownloadResult): { kind: PreviewKind; mime: string } {
  let mime = (file.blob.type.split(';')[0] || '').trim().toLowerCase()
  if (!mime || mime === 'application/octet-stream') {
    const extension = file.filename?.includes('.')
      ? file.filename.split('.').pop()?.toLowerCase() || ''
      : ''
    mime = extensionTypes[extension] || ''
  }
  if (
    mime.startsWith('text/') ||
    textApplications.has(mime) ||
    mime.endsWith('+json') ||
    (mime.endsWith('+xml') && mime !== 'image/svg+xml')
  )
    return { kind: 'text', mime }
  if (images.has(mime)) return { kind: 'image', mime }
  if (audio.has(mime)) return { kind: 'audio', mime }
  if (video.has(mime)) return { kind: 'video', mime }
  return { kind: 'unsupported', mime }
}
function release() {
  if (objectUrl.value) URL.revokeObjectURL(objectUrl.value)
  objectUrl.value = ''
  text.value = ''
  result.value = undefined
}
function close() {
  requestNumber++
  release()
  loading.value = false
  error.value = ''
}
async function load() {
  const request = ++requestNumber
  release()
  loading.value = true
  error.value = ''
  try {
    const file = await api.downloadBackup(projectId.value, versionId.value)
    if (request !== requestNumber) return
    const preview = previewType(file)
    const contents = preview.kind === 'text' ? await file.blob.text() : ''
    if (request !== requestNumber) return
    result.value = file
    kind.value = preview.kind
    text.value = contents
    if (preview.kind !== 'text' && preview.kind !== 'unsupported') {
      objectUrl.value = URL.createObjectURL(file.blob.slice(0, file.blob.size, preview.mime))
    }
  } catch (e) {
    if (request === requestNumber) error.value = errorMessage(e)
  } finally {
    if (request === requestNumber) loading.value = false
  }
}
function open(project: string, version: string) {
  projectId.value = project
  versionId.value = version
  visible.value = true
  void load()
}
function mediaError() {
  error.value = '浏览器无法预览此文件，请下载后查看。'
}
function download() {
  if (result.value) saveFile(result.value.blob, result.value.filename)
}
onBeforeUnmount(close)
defineExpose({ open })
</script>

<template>
  <el-dialog
    v-model="visible"
    title="文件预览"
    width="min(900px, 92vw)"
    destroy-on-close
    @close="close"
  >
    <div class="preview-body" v-loading="loading" aria-live="polite">
      <el-alert v-if="error" :title="error" type="error" show-icon :closable="false" />
      <template v-else-if="!loading && result">
        <p class="preview-filename">{{ result.filename || '文件' }}</p>
        <pre v-if="kind === 'text'" class="preview-text">{{ text }}</pre>
        <img
          v-else-if="kind === 'image'"
          class="preview-image"
          :src="objectUrl"
          :alt="result.filename || '文件预览'"
          @error="mediaError"
        />
        <audio
          v-else-if="kind === 'audio'"
          class="preview-media"
          :src="objectUrl"
          controls
          @error="mediaError"
        />
        <video
          v-else-if="kind === 'video'"
          class="preview-video"
          :src="objectUrl"
          controls
          @error="mediaError"
        />
        <el-empty v-else description="此文件类型暂不支持预览，请下载后查看。" />
      </template>
    </div>
    <template #footer>
      <el-button v-if="error && !result" @click="load">重试</el-button>
      <el-button v-if="result" type="primary" @click="download">下载</el-button>
      <el-button @click="visible = false">关闭</el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.preview-body {
  min-height: 180px;
  max-height: 70vh;
  overflow: auto;
}
.preview-filename {
  margin: 0 0 12px;
  overflow-wrap: anywhere;
  color: var(--el-text-color-secondary);
}
.preview-text {
  margin: 0;
  padding: 16px;
  overflow: auto;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
  background: var(--el-fill-color-light);
}
.preview-image {
  display: block;
  max-width: 100%;
  max-height: 60vh;
  margin: auto;
}
.preview-media {
  width: 100%;
}
.preview-video {
  display: block;
  max-width: 100%;
  max-height: 60vh;
  margin: auto;
}
</style>
