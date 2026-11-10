import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import LoginView from '@/views/LoginView.vue'
import HomeView from '@/views/HomeView.vue'
import AdminLayout from '@/layouts/AdminLayout.vue'
import DashboardView from '@/views/DashboardView.vue'
import ProjectsView from '@/views/ProjectsView.vue'
import ProjectView from '@/views/ProjectView.vue'
import FileDetailView from '@/views/FileDetailView.vue'
import TokensView from '@/views/TokensView.vue'
import SettingsView from '@/views/SettingsView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/', name: 'home', component: HomeView },
    { path: '/login', name: 'login', component: LoginView },
    {
      path: '/',
      component: AdminLayout,
      meta: { requiresAuth: true },
      children: [
        { path: 'dashboard', name: 'dashboard', component: DashboardView, meta: { title: '概览' } },
        { path: 'projects', name: 'projects', component: ProjectsView, meta: { title: '项目' } },
        {
          path: 'projects/:id',
          name: 'project',
          component: ProjectView,
          meta: { title: '项目详情' },
        },
        {
          path: 'projects/:id/files/:fileId',
          name: 'file-detail',
          component: FileDetailView,
          meta: { title: '文件详情' },
        },
        { path: 'tokens', name: 'tokens', component: TokensView, meta: { title: 'API Token' } },
        {
          path: 'settings',
          name: 'settings',
          component: SettingsView,
          meta: { title: '账户设置' },
        },
      ],
    },
  ],
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (!to.meta.requiresAuth) return true
  if (!auth.token) return { name: 'login' }
  try {
    if (await auth.verify()) return true
    return { name: 'login', query: { expired: '1' } }
  } catch {
    return { name: 'login', query: { unavailable: '1' } }
  }
})

export default router
