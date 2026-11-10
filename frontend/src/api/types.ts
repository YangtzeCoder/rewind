export interface Pagination {
  page?: number
  page_size?: number
}

export interface LoginResult {
  token: string
  expires_at: string
}

export interface Project {
  id: string
  name: string
  description: string
  created_at: string
  updated_at: string
}

export interface ProjectInput {
  name: string
  description?: string | null
}

export interface ProjectUpdate {
  name?: string | null
  description?: string | null
}

export interface Dashboard {
  project_count: number
  file_count: number
  version_count: number
  total_size: number
}

export interface ProjectFile {
  id: string
  key: string
  created_at: string
  updated_at: string
  version_count: number
  latest_version: { version_id: string; size: number; created_at: string } | null
}

export type FileDetail = Pick<ProjectFile, 'id' | 'key'>

export interface FileListOptions extends Pagination {
  key?: string
}

export interface ApiToken {
  id: string
  name: string
  created_at: string
}

export interface CreatedApiToken extends ApiToken {
  token: string
}

export interface BackupHistoryRequest extends Pagination {
  project_id: string
  key: string
}

export interface BackupVersion {
  version_id: string
  size: number
  description: string
  created_at: string
}

export interface DownloadResult {
  blob: Blob
  filename: string | undefined
  contentDisposition: string | null
}
