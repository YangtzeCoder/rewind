import type {
  ApiToken, BackupHistoryRequest, BackupVersion, CreatedApiToken, Dashboard, DownloadResult,
  FileListOptions, LoginResult, Pagination, Project, ProjectFile, ProjectInput, ProjectUpdate,
} from './types'

export type * from './types'

export type ApiErrorKind = 'business' | 'network' | 'http' | 'response'

export class ApiError extends Error {
  constructor(
    public readonly kind: ApiErrorKind,
    message: string,
    public readonly code?: string,
    public readonly status?: number,
    options?: ErrorOptions,
  ) {
    super(message, options)
    this.name = 'ApiError'
  }
}

export interface ApiClientOptions {
  /** Backend origin or base URL; defaults to the current origin. */
  baseUrl?: string
  /** Called for every authenticated request so the caller can change tokens. */
  getToken?: () => string | null | undefined
  onUnauthorized?: () => void
}

function query(params?: Pagination | FileListOptions): string {
  const values = new URLSearchParams()
  if (params?.page !== undefined) values.set('page', String(params.page))
  if (params?.page_size !== undefined) values.set('page_size', String(params.page_size))
  if (params && 'key' in params && params.key !== undefined) values.set('key', params.key)
  const result = values.toString()
  return result ? `?${result}` : ''
}

function filenameFromDisposition(value: string | null): string | undefined {
  if (!value) return undefined
  const extended = /(?:^|;)\s*filename\*\s*=\s*UTF-8''([^;]*)/i.exec(value)?.[1]
  if (extended !== undefined) {
    try { return decodeURIComponent(extended) } catch { /* Try the plain filename. */ }
  }
  const plain = /(?:^|;)\s*filename\s*=\s*(?:"((?:\\.|[^"\\])*)"|([^;]*))/i.exec(value)
  return plain?.[1]?.replace(/\\(.)/g, '$1') ?? plain?.[2]?.trim()
}

export function createApiClient(options: ApiClientOptions = {}) {
  const base = (options.baseUrl ?? '').replace(/\/+$/, '') + '/api/v1'

  async function send(path: string, init: RequestInit = {}, authenticated = true): Promise<Response> {
    const headers = new Headers(init.headers)
    if (authenticated) {
      const token = options.getToken?.()
      if (token) headers.set('Authorization', `Bearer ${token}`)
    }
    let response: Response
    try {
      response = await fetch(base + path, { ...init, headers })
    } catch (cause) {
      throw new ApiError('network', 'Network request failed', undefined, undefined, { cause })
    }
    if (!response.ok) throw new ApiError('http', `HTTP ${response.status}`, undefined, response.status)
    return response
  }

  async function parse<T>(response: Response): Promise<T> {
    let body: unknown
    try {
      body = await response.json()
    } catch (cause) {
      throw new ApiError('response', 'Invalid JSON response', undefined, response.status, { cause })
    }
    if (typeof body !== 'object' || body === null || !('success' in body)) {
      throw new ApiError('response', 'Invalid API response', undefined, response.status)
    }
    if (body.success === false) {
      if ('error' in body && typeof body.error === 'object' && body.error !== null &&
          'code' in body.error && typeof body.error.code === 'string' &&
          'message' in body.error && typeof body.error.message === 'string') {
        if (body.error.code === 'UNAUTHORIZED') options.onUnauthorized?.()
        throw new ApiError('business', body.error.message, body.error.code, response.status)
      }
      throw new ApiError('response', 'Invalid API error response', undefined, response.status)
    }
    if (body.success !== true) throw new ApiError('response', 'Invalid API response', undefined, response.status)
    return ('data' in body ? body.data : undefined) as T
  }

  async function request<T>(path: string, init?: RequestInit, authenticated = true): Promise<T> {
    return parse<T>(await send(path, init, authenticated))
  }

  function json(method: string, body: unknown): RequestInit {
    return { method, headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) }
  }

  return {
    login: (username: string, password: string) =>
      request<LoginResult>('/auth/login', json('POST', { username, password }), false),
    logout: () => request<void>('/auth/logout', { method: 'POST' }),
    changePassword: (current_password: string, new_password: string) =>
      request<void>('/auth/password', json('POST', { current_password, new_password })),
    me: () => request<{ username: string }>('/auth/me'),
    dashboard: () => request<Dashboard>('/dashboard'),
    listProjects: (params?: Pagination) => request<Project[]>(`/projects${query(params)}`),
    createProject: (input: ProjectInput) => request<Project>('/projects', json('POST', input)),
    getProject: (id: string) => request<Project>(`/projects/${encodeURIComponent(id)}`),
    updateProject: (id: string, input: ProjectUpdate) =>
      request<Project>(`/projects/${encodeURIComponent(id)}`, json('PATCH', input)),
    deleteProject: (id: string) => request<void>(`/projects/${encodeURIComponent(id)}`, { method: 'DELETE' }),
    listFiles: (projectId: string, params?: FileListOptions) =>
      request<ProjectFile[]>(`/projects/${encodeURIComponent(projectId)}/files${query(params)}`),
    listTokens: (params?: Pagination) => request<ApiToken[]>(`/tokens${query(params)}`),
    createToken: (name: string) => request<CreatedApiToken>('/tokens', json('POST', { name })),
    deleteToken: (id: string) => request<void>(`/tokens/${encodeURIComponent(id)}`, { method: 'DELETE' }),
    backupHistory: (input: BackupHistoryRequest) =>
      request<BackupVersion[]>('/backups/history', json('POST', input)),
    async downloadBackup(projectId: string, versionId: string): Promise<DownloadResult> {
      const response = await send(`/backups/download/${encodeURIComponent(projectId)}/${encodeURIComponent(versionId)}`)
      const disposition = response.headers.get('Content-Disposition')
      // A downloaded file can itself be JSON; successful downloads have an attachment header.
      if (!disposition && response.headers.get('Content-Type')?.toLowerCase().includes('application/json')) {
        await parse<never>(response)
        throw new ApiError('response', 'Expected a file download', undefined, response.status)
      }
      if (!disposition) throw new ApiError('response', 'Missing download disposition', undefined, response.status)
      try {
        return { blob: await response.blob(), filename: filenameFromDisposition(disposition), contentDisposition: disposition }
      } catch (cause) {
        throw new ApiError('network', 'Failed to read download', undefined, response.status, { cause })
      }
    },
  }
}
