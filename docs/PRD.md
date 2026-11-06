# AI Agent 文档备份服务 PRD

**版本：** v1.0
**产品定位：** 面向 AI Agent 的开源、自托管文件备份与版本管理服务。

## 1. 功能范围

### 1.1 项目管理

- 创建、查看、编辑、删除项目。
- 每个项目独立管理文件和备份版本。
- 项目使用系统生成的唯一 `project_id`。

### 1.2 文件备份

- 通过 REST API 上传单个文件，每次请求仅处理一个文件。
- 请求必须包含 `project_id`、`key` 和文件内容。
- 支持可选 `description`，最长 1024 字符。
- 同一项目下，相同 `key` 的多次备份形成版本历史。
- 每次备份均独立保存一份文件内容，即使内容与已有版本相同也不去重。
- `key` 由客户端指定，通常为相对路径，允许较长字符串及特殊字符。
- 文件通过 `(project_id, key)` 定位。

### 1.3 版本管理

- 查询指定文件的版本历史。
- 展示版本 ID、创建时间、文件大小及描述。
- 下载指定历史版本。
- 支持下载最新版本。

本阶段不实现保留策略或基于保留策略的自动清理，备份版本不会因数量或保存时间被自动删除。

## 2. 管理后台

提供 Web 管理界面，包含：

- 管理员使用固定用户名 `admin` 登录、退出和修改密码。
- 系统概览。
- 项目管理。
- 文件与版本历史查看、下载。
- API Token 创建、列表查看及删除。

系统仅提供固定用户名 `admin` 的管理员账号，不提供账号管理、角色或权限配置。

## 3. 认证与权限

### 3.1 管理员账号

- 系统仅提供固定用户名 `admin`，密码通过部署配置提供；不提供默认密码，首次启动必须配置密码。
- 密码只保存安全哈希，日志不记录密码。

### 3.2 登录 Token

- 管理员网页登录成功后签发登录 Token；后台在后续请求的 `Authorization: Bearer <token>` 中携带它。登录 Token 的闲置有效期为 30 分钟：从签发或最近一次成功通过认证的请求开始计算，连续 30 分钟未使用即失效；每次成功通过认证时将失效时间续至当前时间起 30 分钟。
- 失效后须重新登录；退出使当前登录 Token 失效。修改管理员密码不影响已有登录 Token，新密码在后续登录时生效。
- 登录 Token 保存在独立的数据表中，记录安全哈希、创建时间、过期时间和失效状态等信息，不保存明文。

### 3.3 API Token

- 管理员可在后台手动创建、查看和删除供 Agent 使用的全局 API Token；创建时提供名称。API Token 无自动过期时间，持续有效直至手动删除；退出登录和修改密码不影响已有 API Token。Token 明文仅在创建时展示一次，服务端仅保存安全哈希；列表展示名称、标识和创建时间，不展示明文。删除后立即失效。日志不记录完整 Token。
- Agent 在每个备份、历史查询或下载请求中通过 `Authorization: Bearer <token>` 携带 API Token，不需要登录或浏览器 Cookie。API Token 不绑定项目、不配置角色或权限；有效 Token 可访问所有项目的 Agent 接口，应妥善保管。
- API Token 保存在独立的数据表中，与登录 Token 使用可区分的格式；日志不记录任何完整 Token。

### 3.4 权限范围

- 本阶段不实现角色、项目级授权或其他可配置权限管理；已登录管理员可操作所有项目，API Token 可访问所有项目的 Agent 接口。
- 服务端按 Token 类型、有效性和接口用途校验：后台管理接口使用登录 Token；备份上传接口仅接受 API Token；历史查询及下载接口接受 API Token，后台也可用登录 Token 调用历史查询及下载接口。


## 4. REST API

通用请求和响应约定见 [API 设计规范](API_DESIGN.md)。

统一使用 `/api/v1` 前缀。登录接口无需 Token；管理后台接口及 Token 管理接口使用登录 Token；备份上传接口仅接受 API Token；历史查询和下载接口允许使用 API Token，也允许已登录后台使用登录 Token。两类 Token 均通过 `Authorization: Bearer <token>` 传递；不做项目级权限校验。删除 API Token 只影响该 Token，不影响管理员登录 Token。

JSON 响应统一使用 HTTP 200，包含表示操作是否成功的 `success`。成功且有返回内容时包含 `data`；无返回内容时省略 `data`。返回数组的接口始终返回数组类型的 `data`，无结果时返回空数组 `[]`。失败时包含 `error` 对象，其中 `code` 为字符串错误码、`message` 为错误说明，并省略 `data`；成功时省略 `error`。文件下载成功时返回原始二进制内容，失败时返回上述 JSON 错误。

错误码以[代码中的定义](../backend/internal/server/response.go)为准。内部错误不向客户端暴露实现细节。

### 认证接口

| 方法 | 路径 | 请求参数 | 功能 |
| --- | --- | --- | --- |
| POST | `/api/v1/auth/login` | `username`, `password` | 管理员登录，返回登录 Token 和初始过期时间；后续使用时服务端按闲置规则续期 |
| POST | `/api/v1/auth/logout` | 无 | 使当前登录 Token 失效 |
| POST | `/api/v1/auth/password` | `current_password`, `new_password` | 修改管理员密码 |
| GET | `/api/v1/auth/me` | 无 | 查询当前登录状态 |

### API Token 管理接口

| 方法 | 路径 | 请求参数 | 功能 |
| --- | --- | --- | --- |
| POST | `/api/v1/tokens` | `name` | 创建永久有效的全局 API Token，仅本次返回明文 |
| GET | `/api/v1/tokens` | 分页参数 | 查询 Token 列表，不返回明文 |
| DELETE | `/api/v1/tokens/{token_id}` | 无 | 删除 Token，使其立即失效 |

### 备份与版本接口

| 方法 | 路径 | 请求参数 | 功能 |
| --- | --- | --- | --- |
| POST | `/api/v1/backups` | `project_id`, `key`, `file`, `description?` | 上传文件备份 |
| POST | `/api/v1/backups/history` | `project_id`, `key`, `page?`, `page_size?` | 查询版本历史 |
| GET | `/api/v1/backups/download/{project_id}/{version_id}` | 路径中的项目 ID、版本 ID | 下载指定版本 |

上传使用 `multipart/form-data`；历史查询使用 JSON Body，`key` 不放入 URL 路径；历史记录按创建顺序倒序，下载最新版本可使用第一条记录的版本 ID。下载使用 GET，返回文件二进制内容。

### 管理后台接口

| 方法 | 路径 | 请求参数 | 功能 |
| --- | --- | --- | --- |
| GET | `/api/v1/dashboard` | 无 | 查询系统概览统计 |
| GET | `/api/v1/projects` | 分页参数 | 查询项目列表 |
| POST | `/api/v1/projects` | 项目信息 | 创建项目 |
| GET | `/api/v1/projects/{project_id}` | 无 | 查看项目详情 |
| PATCH | `/api/v1/projects/{project_id}` | 待修改的项目信息 | 编辑项目 |
| DELETE | `/api/v1/projects/{project_id}` | 无 | 删除项目及其备份数据 |
| GET | `/api/v1/projects/{project_id}/files` | 分页、筛选参数 | 查询项目文件列表 |

后台通过上述备份历史及下载接口查看、下载版本。项目删除前须在后台明确提示将删除该项目的文件和版本，并要求用户确认；删除项目时一并删除其各版本独立保存的文件。列表接口支持分页，概览及文件列表的字段和筛选规则见 API 设计规范。

## 5. ID 规范

普通实体 ID 采用类型前缀和 16 位随机字符。

| 实体          | 格式                            |
| ------------- | ------------------------------- |
| Project       | `PRG_<16位随机ID>`              |
| File          | `FILE_<16位随机ID>`             |
| Version       | `VER_<16位随机ID>`              |
| API Token     | `TKN_<16位随机ID>`              |

API Token 的完整凭证包含独立的安全随机密钥，具体编码由接口设计确定。

File ID 为内部标识，Agent API 使用 `project_id` 和 `key` 定位文件。

## 6. 存储与部署

### 6.1 存储

- 每个版本独立保存文件内容，使用版本 ID 定位对应的存储对象，不进行内容去重或引用计数。
- 使用 SHA256 校验文件完整性，不作为去重依据。
- 当前实现本地文件系统存储后端。
- 存储层采用统一接口，与业务逻辑解耦。
- 存储接口覆盖版本文件的写入、读取、存在性检查及删除。
- 存储接口设计支持后续扩展 S3 兼容对象存储及其他存储后端。
- 元数据使用 SQLite 存储。
- 支持流式上传、校验及原子写入。

### 6.2 部署

- 使用 Docker 单容器部署。
- 数据通过 Docker Volume 持久化。

命令行支持：

| 命令或参数 | 说明 |
| --- | --- |
| `rewind serve` | 启动服务 |
| `--listen` | 监听地址，默认 `0.0.0.0:8080` |
| `--data-dir` | 数据目录，默认 `/data` |
| `rewind version` | 查看版本 |
| `rewind --help` | 查看帮助 |

`--listen` 和 `--data-dir` 也可分别通过环境变量 `REWIND_LISTEN` 和 `REWIND_DATA_DIR` 配置；命令行参数优先于环境变量。

首次初始化时，通过 `REWIND_ADMIN_PASSWORD` 或 `REWIND_ADMIN_PASSWORD_FILE` 提供管理员初始密码，不通过命令行参数传入。初始密码仅在管理员账户尚未创建时使用；首次启动未提供密码则报错退出。之后更改环境变量不会覆盖已有密码。

## 7. 技术栈

| 模块           | 技术                 |
| -------------- | -------------------- |
| 后端           | Go + Gin             |
| 数据库         | SQLite               |
| 文件存储       | 独立版本文件 + 可扩展存储接口 |
| 默认存储后端   | 本地文件系统         |
| 前端           | Vue 3 + TypeScript   |
| UI             | Element Plus         |
| 构建及状态管理 | Vite + Pinia         |
| 部署           | Docker               |
| 前端资源       | Go Embed             |
