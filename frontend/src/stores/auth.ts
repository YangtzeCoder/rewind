import { ref } from 'vue'
import { defineStore } from 'pinia'
import { ApiError, createApiClient } from '@/api'

const tokenKey = 'rewind.login-token'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(sessionStorage.getItem(tokenKey))
  const username = ref('')
  const api = createApiClient({
    getToken: () => token.value,
    onUnauthorized: () => {
      clear()
      window.dispatchEvent(new Event('rewind:unauthorized'))
    },
  })

  function clear() {
    token.value = null
    username.value = ''
    sessionStorage.removeItem(tokenKey)
  }

  async function login(name: string, password: string) {
    const result = await api.login(name, password)
    token.value = result.token
    username.value = name
    sessionStorage.setItem(tokenKey, result.token)
  }

  async function verify() {
    if (!token.value) return false
    try {
      username.value = (await api.me()).username
      return true
    } catch (error) {
      if (error instanceof ApiError && error.code === 'UNAUTHORIZED') {
        clear()
        return false
      }
      throw error
    }
  }

  async function logout() {
    try {
      await api.logout()
    } catch (error) {
      if (!(error instanceof ApiError && error.code === 'UNAUTHORIZED')) throw error
    }
    clear()
  }

  return { token, username, api, login, verify, logout }
})
