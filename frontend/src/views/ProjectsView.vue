<script setup lang="ts">
import { onMounted, ref } from 'vue'
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
const name = ref('')
const description = ref('')
async function load() {
  loading.value = true; error.value = ''; projects.value = []
  try { projects.value = await api.listProjects({ page: page.value, page_size: 20 }) } catch (e) { error.value = errorMessage(e) } finally { loading.value = false }
}
function openForm(project?: Project) {
  editing.value = project ?? null; name.value = project?.name ?? ''; description.value = project?.description ?? ''; showForm.value = true; error.value = ''
}
async function save() {
  if (!name.value.trim() || saving.value) return
  saving.value = true; error.value = ''
  try {
    if (editing.value) await api.updateProject(editing.value.id, { name: name.value, description: description.value })
    else { await api.createProject({ name: name.value, description: description.value }); page.value = 1 }
    showForm.value = false; await load()
  } catch (e) { error.value = errorMessage(e) } finally { saving.value = false }
}
async function remove(project: Project) {
  if (!window.confirm(`确定删除项目“${project.name}”吗？该项目的所有文件和备份版本也将被永久删除。`)) return
  try { await api.deleteProject(project.id); if (projects.value.length === 1 && page.value > 1) page.value--; await load() } catch (e) { error.value = errorMessage(e) }
}
function changePage(delta: number) { page.value += delta; void load() }
onMounted(load)
</script>
<template>
  <section class="dashboard">
    <div class="page-heading"><div><div class="eyebrow">WORKSPACE</div><h1>项目</h1><p>管理项目及其中保存的备份。</p></div><el-button type="primary" @click="openForm()">新建项目</el-button></div>
    <p v-if="error" class="form-error" role="alert">{{ error }} <button v-if="!showForm" class="text-button" @click="load">重试</button></p>
    <form v-if="showForm" class="panel form-panel" @submit.prevent="save">
      <div class="panel-title"><h2>{{ editing ? '编辑项目' : '新建项目' }}</h2><button type="button" class="text-button" @click="showForm = false">关闭</button></div>
      <label for="project-name">项目名称</label><el-input id="project-name" v-model="name" placeholder="输入项目名称" />
      <label for="project-description">描述（可选）</label><el-input id="project-description" v-model="description" type="textarea" :rows="3" placeholder="为项目添加描述" />
      <div class="form-actions"><el-button @click="showForm = false">取消</el-button><el-button type="primary" native-type="submit" :loading="saving" :disabled="!name.trim()">保存</el-button></div>
    </form>
    <div class="panel list-panel">
      <p v-if="loading" class="empty-text">正在加载…</p><p v-else-if="!projects.length" class="empty-text">{{ page === 1 ? '还没有项目，创建一个项目开始备份。' : '没有更多项目。' }}</p>
      <div v-for="project in projects" v-else :key="project.id" class="list-row">
        <div class="row-main"><RouterLink class="item-title" :to="{ name: 'project', params: { id: project.id } }">{{ project.name }}</RouterLink><p>{{ project.description || '暂无描述' }}</p><small>{{ project.id }} · 创建于 {{ formatDate(project.created_at) }}</small></div>
        <div class="row-actions"><button class="text-button" @click="openForm(project)">编辑</button><button class="text-button danger" @click="remove(project)">删除</button></div>
      </div>
      <div class="pagination"><el-button :disabled="page === 1 || loading" @click="changePage(-1)">上一页</el-button><span>第 {{ page }} 页</span><el-button :disabled="projects.length < 20 || loading" @click="changePage(1)">下一页</el-button></div>
    </div>
  </section>
</template>
