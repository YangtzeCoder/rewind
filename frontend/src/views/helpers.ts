import { ApiError } from '@/api'

export function errorMessage(error: unknown): string {
  if (error instanceof ApiError) {
    if (error.code === 'INVALID_CREDENTIALS') return '当前密码不正确。'
    if (error.code === 'NOT_FOUND') return '请求的内容不存在或已被删除。'
    if (error.code === 'INVALID_ARGUMENT') return '输入内容无效，请检查后重试。'
    if (error.code === 'UNAUTHORIZED') return '登录已过期，请重新登录。'
  }
  return '操作失败，请检查网络或稍后重试。'
}

export function formatDate(date: string): string {
  return new Date(date).toLocaleString('zh-CN')
}

export function formatSize(size: number): string {
  if (size < 1024) return `${size} B`
  const units = ['KB', 'MB', 'GB', 'TB']
  let value = size / 1024
  let unit = 0
  while (value >= 1024 && unit < units.length - 1) {
    value /= 1024
    unit++
  }
  return `${value.toFixed(1)} ${units[unit]}`
}

export function saveFile(blob: Blob, filename: string | undefined) {
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename || 'download'
  link.click()
  setTimeout(() => URL.revokeObjectURL(url), 1000)
}
