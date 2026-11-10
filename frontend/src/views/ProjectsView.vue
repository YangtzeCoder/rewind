<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessageBox } from 'element-plus'
import type { Project } from '@/api'
import { useAuthStore } from '@/stores/auth'
import { errorMessage, formatDate } from './helpers'

const api = useAuthStore().api
const projects = ref<Project[]>([])
const page = ref(1)
const loading = ref(false)
const saving = ref(false)
const error = ref('')
const editing = ref<Project | null>(null)
const showForm = ref(false)
const formError = ref('')
const name = ref('')
const description = ref('')
const copiedId = ref('')
async function copyId(id: string) {
  try {
    await navigator.clipboard.writeText(id)
    copiedId.value = id
  } catch {
    error.value = '复制失败，请手动选择并复制项目 ID。'
  }
}
async function load() {
  loading.value = true
  error.value = ''
  projects.value = []
  try {
    projects.value = await api.listProjects({ page: page.value, page_size: 20 })
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    loading.value = false
  }
}
function openForm(project?: Project) {
  editing.value = project ?? null
  name.value = project?.name ?? ''
  description.value = project?.description ?? ''
  showForm.value = true
  formError.value = ''
}
async function save() {
  if (!name.value.trim() || saving.value) return
  saving.value = true
  formError.value = ''
  try {
    if (editing.value)
      await api.updateProject(editing.value.id, {
        name: name.value,
        description: description.value,
      })
    else {
      await api.createProject({ name: name.value, description: description.value })
      page.value = 1
    }
    showForm.value = false
    await load()
  } catch (e) {
    formError.value = errorMessage(e)
  } finally {
    saving.value = false
  }
}
async function remove(project: Project) {
  try {
    await ElMessageBox.confirm(
      `确定删除项目“${project.name}”吗？该项目的所有文件和备份版本也将被永久删除。`,
      '删除项目',
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
    await api.deleteProject(project.id)
    if (projects.value.length === 1 && page.value > 1) page.value--
    await load()
  } catch (e) {
    error.value = errorMessage(e)
  }
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
        <h1>项目</h1>
        <p>管理项目及其中保存的备份。</p>
      </div>
      <el-button type="primary" @click="openForm()">新建项目</el-button>
    </div>
    <el-alert v-if="error" :title="error" type="error" show-icon :closable="false">
      <el-button link @click="load">重试</el-button>
    </el-alert>
    <el-dialog
      v-model="showForm"
      :title="editing ? '编辑项目' : '新建项目'"
      width="min(90%, 560px)"
      :close-on-click-modal="false"
      :close-on-press-escape="!saving"
      :show-close="!saving"
    >
      <el-alert v-if="formError" :title="formError" type="error" show-icon :closable="false" />
      <el-form label-position="top" @submit.prevent="save">
        <el-form-item label="项目名称" required>
          <el-input v-model="name" placeholder="输入项目名称" />
        </el-form-item>
        <el-form-item label="描述（可选）">
          <el-input v-model="description" type="textarea" :rows="3" placeholder="为项目添加描述" />
        </el-form-item>
        <div class="dialog-actions">
          <el-button :disabled="saving" @click="showForm = false">取消</el-button>
          <el-button type="primary" native-type="submit" :loading="saving" :disabled="!name.trim()">
            保存
          </el-button>
        </div>
      </el-form>
    </el-dialog>
    <el-card class="content-card">
      <el-table v-loading="loading" :data="projects" style="width: 100%">
        <el-table-column label="项目" min-width="240">
          <template #default="{ row }">
            <el-link
              type="primary"
              :underline="false"
              @click="$router.push({ name: 'project', params: { id: row.id } })"
            >
              {{ row.name }}
            </el-link>
            <div class="secondary-text">{{ row.description || '暂无描述' }}</div>
            <div class="secondary-text break-text">
              {{ row.id }}
              <el-button link type="primary" @click="copyId(row.id)">
                {{ copiedId === row.id ? '已复制' : '复制' }}
              </el-button>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" min-width="170">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="140">
          <template #default="{ row }">
            <el-button link type="primary" @click="openForm(row)">编辑</el-button>
            <el-button link type="danger" @click="remove(row)">删除</el-button>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty
            :description="page === 1 ? '还没有项目，创建一个项目开始备份。' : '没有更多项目。'"
          />
        </template>
      </el-table>
      <div class="pagination">
        <el-button :disabled="page === 1 || loading" @click="changePage(-1)">上一页</el-button>
        <span>第 {{ page }} 页</span>
        <el-button :disabled="projects.length < 20 || loading" @click="changePage(1)">
          下一页
        </el-button>
      </div>
    </el-card>
  </section>
</template>
