<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessageBox } from 'element-plus'
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
  loading.value = true
  error.value = ''
  tokens.value = []
  try {
    tokens.value = await api.listTokens({ page: page.value, page_size: 20 })
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    loading.value = false
  }
}
async function create() {
  if (!name.value.trim() || creating.value || createdToken.value) return
  creating.value = true
  error.value = ''
  try {
    const result = await api.createToken(name.value)
    createdToken.value = result.token
    name.value = ''
    page.value = 1
    await load()
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    creating.value = false
  }
}
async function remove(token: ApiToken) {
  try {
    await ElMessageBox.confirm(
      `确定删除 Token“${token.name}”吗？使用它的客户端将立即失去访问权限。`,
      '删除 Token',
      {
        type: 'warning',
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        closeOnClickModal: false,
      },
    )
  } catch {
    return
  }
  try {
    await api.deleteToken(token.id)
    if (tokens.value.length === 1 && page.value > 1) page.value--
    await load()
  } catch (e) {
    error.value = errorMessage(e)
  }
}
async function copy() {
  try {
    await navigator.clipboard.writeText(createdToken.value)
    copied.value = true
  } catch {
    error.value = '复制失败，请手动选择并复制 Token。'
  }
}
function dismissCreatedToken() {
  createdToken.value = ''
  copied.value = false
}
function changePage(delta: number) {
  page.value += delta
  void load()
}
onMounted(load)
</script>
<template>
  <section class="dashboard">
    <div class="page-heading">
      <div>
        <h1>API Token</h1>
        <p>创建供 Agent 使用的访问凭证。</p>
      </div>
    </div>
    <el-alert v-if="error" :title="error" type="error" show-icon :closable="false">
      <el-button link @click="load">重试加载</el-button>
    </el-alert>
    <el-card v-if="createdToken" class="content-card">
      <el-alert
        title="请立即保存 Token"
        description="完整 Token 仅在创建时显示一次，关闭后无法再次查看。"
        type="warning"
        show-icon
        :closable="false"
      />
      <div class="token-value">
        <code>{{ createdToken }}</code>
        <el-button @click="copy">{{ copied ? '已复制' : '复制' }}</el-button>
      </div>
      <el-button @click="dismissCreatedToken">已保存，关闭</el-button>
    </el-card>
    <el-card class="content-card">
      <template #header>创建 Token</template>
      <el-form class="token-form" @submit.prevent="create">
        <el-form-item label="Token 名称">
          <el-input v-model="name" aria-label="Token 名称" placeholder="例如：我的 Agent" />
        </el-form-item>
        <el-button
          type="primary"
          native-type="submit"
          :loading="creating"
          :disabled="!name.trim() || !!createdToken"
        >
          创建
        </el-button>
      </el-form>
    </el-card>
    <el-card class="content-card">
      <template #header>已创建的 Token</template>
      <el-table v-loading="loading" :data="tokens">
        <el-table-column prop="name" label="名称" min-width="150" />
        <el-table-column prop="id" label="ID" min-width="220" />
        <el-table-column label="创建时间" min-width="170">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="90">
          <template #default="{ row }">
            <el-button link type="danger" @click="remove(row)">删除</el-button>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty :description="page === 1 ? '还没有 API Token。' : '没有更多 Token。'" />
        </template>
      </el-table>
      <div class="pagination">
        <el-button :disabled="page === 1 || loading" @click="changePage(-1)">上一页</el-button>
        <span>第 {{ page }} 页</span>
        <el-button :disabled="tokens.length < 20 || loading" @click="changePage(1)">
          下一页
        </el-button>
      </div>
    </el-card>
  </section>
</template>
