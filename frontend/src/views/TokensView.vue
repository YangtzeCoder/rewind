<script setup lang="ts">
import { onMounted, ref } from 'vue'
import type { ApiToken } from '@/api'
import { useAuthStore } from '@/stores/auth'
import { errorMessage, formatDate } from './helpers'

const api = useAuthStore().api
const tokens = ref<ApiToken[]>([])
const page = ref(1)
const name = ref('')
const createdToken = ref('')
const creating = ref(false)
const loading = ref(false)
const error = ref('')
const copied = ref(false)
async function load() {
  loading.value = true; error.value = ''; tokens.value = []
  try { tokens.value = await api.listTokens({ page: page.value, page_size: 20 }) } catch (e) { error.value = errorMessage(e) } finally { loading.value = false }
}
async function create() {
  if (!name.value.trim() || creating.value || createdToken.value) return
  creating.value = true; error.value = ''
  try { const result = await api.createToken(name.value); createdToken.value = result.token; name.value = ''; page.value = 1; await load() } catch (e) { error.value = errorMessage(e) } finally { creating.value = false }
}
async function remove(token: ApiToken) {
  if (!window.confirm(`确定删除 Token“${token.name}”吗？使用它的客户端将立即失去访问权限。`)) return
  try { await api.deleteToken(token.id); if (tokens.value.length === 1 && page.value > 1) page.value--; await load() } catch (e) { error.value = errorMessage(e) }
}
async function copy() {
  try { await navigator.clipboard.writeText(createdToken.value); copied.value = true } catch { error.value = '复制失败，请手动选择并复制 Token。' }
}
function changePage(delta: number) { page.value += delta; void load() }
onMounted(load)
</script>
<template>
  <section class="dashboard">
    <div class="page-heading"><div><div class="eyebrow">ACCESS</div><h1>API Token</h1><p>创建供 Agent 使用的访问凭证。</p></div></div>
    <p v-if="error" class="form-error" role="alert">{{ error }}</p>
    <div v-if="createdToken" class="token-reveal panel" role="status"><h2>请立即保存 Token</h2><p>完整 Token 仅在创建时显示一次，关闭后无法再次查看。</p><div class="token-value"><code>{{ createdToken }}</code><el-button @click="copy">{{ copied ? '已复制' : '复制' }}</el-button></div><el-button @click="createdToken = ''; copied = false">已保存，关闭</el-button></div>
    <form class="panel inline-form" @submit.prevent="create"><div><h2>创建 Token</h2><p class="muted">为 Token 起一个便于识别的名称。</p></div><el-input v-model="name" aria-label="Token 名称" placeholder="例如：我的 Agent" /><el-button type="primary" native-type="submit" :loading="creating" :disabled="!name.trim() || !!createdToken">创建</el-button></form>
    <div class="panel list-panel"><div class="panel-title"><h2>已创建的 Token</h2></div><p v-if="loading" class="empty-text">正在加载…</p><p v-else-if="!tokens.length" class="empty-text">{{ page === 1 ? '还没有 API Token。' : '没有更多 Token。' }}</p>
      <div v-for="token in tokens" v-else :key="token.id" class="list-row"><div class="row-main"><strong class="item-title">{{ token.name }}</strong><p>{{ token.id }}</p><small>创建于 {{ formatDate(token.created_at) }}</small></div><button class="text-button danger" @click="remove(token)">删除</button></div>
      <div class="pagination"><el-button :disabled="page === 1 || loading" @click="changePage(-1)">上一页</el-button><span>第 {{ page }} 页</span><el-button :disabled="tokens.length < 20 || loading" @click="changePage(1)">下一页</el-button></div>
    </div>
  </section>
</template>
