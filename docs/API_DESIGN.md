# API 设计规范

本规范描述 Rewind 接口的通用约定。具体接口及功能范围见 [PRD](PRD.md)。

## 请求

- REST API 统一使用 `/api/v1` 前缀。
- 需要认证的接口通过 `Authorization: Bearer <token>` 传递 Token；各接口允许的 Token 类型见 PRD。
- 文件上传使用 `multipart/form-data`，备份历史查询使用 JSON 请求体，下载使用 GET 路径参数。

## JSON 响应

所有 JSON 响应均使用 HTTP 200。客户端通过 `success` 判断操作是否成功，而不是仅根据 HTTP 状态码判断。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `success` | 布尔值 | 必有；表示操作是否成功。 |
| `data` | 任意 JSON 值 | 成功且有业务数据时返回；没有返回内容时省略。 |
| `error` | 对象 | 失败时返回，包含字符串字段 `code` 和 `message`；成功时省略。 |

有返回内容的成功响应：

```json
{
  "success": true,
  "data": {"username": "admin"}
}
```

没有返回内容的成功响应：

```json
{"success": true}
```

返回数组的接口始终包含数组类型的 `data`。没有结果时返回空数组，不返回 `null` 或省略 `data`：

```json
{"success": true, "data": []}
```

失败响应省略 `data`，通过 `error.code` 标识错误类型，`error.message` 描述原因：

```json
{
  "success": false,
  "error": {
    "code": "<错误码>",
    "message": "<错误说明>"
  }
}
```

`error.code` 和 `error.message` 的默认值均为空字符串；发生错误时填写。错误说明使用英文；客户端应通过稳定的错误码判断错误类型。错误码以[代码中的定义](../backend/internal/server/response.go)为准。内部错误只返回通用说明，不暴露实现细节。

## 文件下载

文件下载成功时直接返回原始二进制内容，不使用 JSON 包装；下载失败时返回统一的 JSON 错误响应。响应内容已经开始发送后，无法再替换为 JSON 错误响应。

`GET /api/v1/backups/download/{project_id}/{version_id}` 接受 API Token 或管理员登录 Token，通过 `Authorization: Bearer <token>` 传递，不接受 URL 参数中的 Token。版本必须属于指定项目。成功时以附件形式返回内容，响应头提供保存的文件名、内容类型和大小；项目或版本不存在返回 `NOT_FOUND`。

## API Token 管理

Token 管理接口只接受登录 Token。`POST /api/v1/tokens` 接收非空 `name`，返回 `id`、`name`、`created_at` 和 `token`；完整 API Token 由 `API_` 前缀和 50 位数字或小写字母组成，仅在创建时返回一次。`id` 是独立的管理标识，不包含在 API Token 中。

`GET /api/v1/tokens` 返回包含 `id`、`name` 和 `created_at` 的数组，不返回 Token 或其哈希。查询参数 `page` 和 `page_size` 均为正整数，默认分别为 1 和 20，`page_size` 最大为 100；按创建时间倒序、ID 倒序排列。`DELETE /api/v1/tokens/{token_id}` 删除指定 Token，不存在时返回 `NOT_FOUND`。时间采用 UTC RFC 3339 格式。

## 项目管理

项目管理接口只接受登录 Token。项目对象包含 `id`、`name`、`description`、`created_at` 和 `updated_at`，时间采用 UTC RFC 3339 格式。名称不要求唯一，名称和描述保留原始输入。

- `POST /api/v1/projects` 接收必填的非空 `name` 和可选的 `description`（缺失或为 `null` 时默认为空字符串），返回创建的项目。名称不能全为空白字符。
- `GET /api/v1/projects` 返回项目数组；`page` 和 `page_size` 均为正整数，默认分别为 1 和 20，`page_size` 最大为 100。结果按创建时间倒序、ID 倒序排列。
- `GET /api/v1/projects/{project_id}` 返回项目详情。
- `PATCH /api/v1/projects/{project_id}` 接收 `name` 和 `description` 中至少一个字段，返回更新后的项目。缺失或为 `null` 的字段不更新，至少一个字段须提供字符串值；名称不能全为空白字符，描述可设为空字符串。
- `DELETE /api/v1/projects/{project_id}` 删除项目及其关联的文件、版本元数据和版本文件。

请求字段无效返回 `INVALID_ARGUMENT`；详情、编辑和删除的项目不存在时返回 `NOT_FOUND`。

### 系统概览

`GET /api/v1/dashboard` 只接受登录 Token。返回 `project_count`、`file_count`、`version_count` 和 `total_size`（所有版本文件大小之和，单位为字节）；没有数据时均为 0。同一文件的各版本分别计入版本数量和总大小。

### 项目文件列表

`GET /api/v1/projects/{project_id}/files` 只接受登录 Token，返回文件数组。查询参数 `page` 和 `page_size` 均为正整数，默认分别为 1 和 20，`page_size` 最大为 100。可选的 `key` 是区分大小写的字面子串筛选，空值表示不筛选，特殊字符按字面匹配。结果按文件更新时间倒序、文件 ID 倒序排列。

每条记录包含 `id`（文件 ID）、`key`、`created_at`、`updated_at`、`version_count` 和 `latest_version`。最新版本按版本创建顺序确定，包含 `version_id`、`size`（字节）和 `created_at`；文件没有版本时 `latest_version` 为 `null`。时间采用 UTC RFC 3339 格式。项目不存在返回 `NOT_FOUND`；分页参数无效或重复的 `key` 返回 `INVALID_ARGUMENT`；没有结果时返回空数组。

## 备份上传

`POST /api/v1/backups` 仅接受 API Token，通过 `Authorization: Bearer <api_token>` 传递；缺少、无效、已删除的 API Token 或使用登录 Token 时返回 `UNAUTHORIZED`。

请求使用 `multipart/form-data`，字段可按任意顺序传递：

| 字段 | 说明 |
| --- | --- |
| `project_id` | 必填，已有项目 ID。 |
| `key` | 必填、非空，保留原始输入，允许特殊字符；与项目 ID 共同定位文件，不作为磁盘路径。 |
| `file` | 必填，恰好一个带文件名的文件，允许空文件，文件内容流式写入。 |
| `description` | 可选，默认为空字符串，最多 1024 个 Unicode 字符。 |

文本字段须为 UTF-8，每个字段最多 1 MiB；不接受重复字段或未定义字段。同一项目、同一 `key` 的上传复用 `file_id`，每次生成新的 `version_id` 并独立保存文件，即使内容相同也不去重。

成功返回 `project_id`、`key`、`file_id`、`version_id`、`size`（字节数）、`sha256`（小写十六进制）、`description`、`original_filename`、`content_type` 和 `created_at`（UTC RFC 3339）。文件名取上传文件名的基础名称；`content_type` 保存文件部分的 Content-Type，未提供时使用 `application/octet-stream`。

项目不存在返回 `NOT_FOUND`，请求格式或字段无效返回 `INVALID_ARGUMENT`。处理失败时清理暂存文件；文件提交后元数据事务失败时清理已提交的文件。

## 备份历史

`POST /api/v1/backups/history` 接受 API Token 或管理员登录 Token，通过 `Authorization: Bearer <token>` 传递。JSON 请求体包含必填的非空 `project_id` 和 `key`，以及可选的正整数 `page` 和 `page_size`，默认值分别为 1 和 20，`page_size` 最大为 100。

成功返回数组，每条记录包含 `version_id`、`size`（字节数）、`description` 和 `created_at`（UTC RFC 3339）。版本按创建顺序倒序，第一页的第一条为最新版本；没有更多结果时返回空数组。客户端根据项目 ID 和版本 ID 构造下载路径，响应不包含下载地址。

项目或文件不存在返回 `NOT_FOUND`；请求格式或字段无效返回 `INVALID_ARGUMENT`。
