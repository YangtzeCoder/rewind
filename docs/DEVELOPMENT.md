# Rewind 开发指南

本文档描述项目的技术结构、本地开发、验证和发布流程。项目行为、REST 接口、字段和错误码以当前代码为准；不要以任何早期需求或 API 草稿覆盖现有实现。

## 技术栈与运行模型

- 后端：Go、Gin、Cobra、SQLite。
- 前端：Vue 3、TypeScript、Vite、Pinia、Element Plus、Vue Router。
- 存储：SQLite 保存元数据；本地文件系统保存每个版本的独立文件。
- 部署：单个 Go 进程提供 REST API 和管理页面，前端资源通过 Go Embed 嵌入。
- 发布：`scripts/build.sh` 构建前端并交叉编译多个平台的二进制文件；GitHub Actions 构建并发布镜像和 GitHub Release。

后端启动时依次：

1. 打开或初始化 `<data-dir>/rewind.db`。
2. 初始化固定管理员账号和数据库结构。
3. 创建本地存储后端。
4. 注册 `/api/v1` 路由。
5. 将前端资源或 SPA 回退交给 HTTP 服务。

## 目录职责

```text
.
├── backend/
│   ├── cmd/rewind/          # 程序入口
│   ├── internal/
│   │   ├── cli/             # Cobra 命令、参数和环境变量
│   │   ├── server/          # Gin 路由、HTTP Handler、认证中间件和响应
│   │   ├── auth/            # 管理员密码、登录会话和认证服务
│   │   ├── database/        # SQLite 初始化和各业务数据访问层
│   │   ├── storage/         # 版本文件存储接口及本地文件实现
│   │   └── id/              # 实体和 Token ID 生成
│   └── http/                # 可用于手工请求的 HTTP 示例
├── frontend/src/
│   ├── api/                 # HTTP 客户端及响应类型
│   ├── components/          # 复用组件
│   ├── layouts/             # 管理后台布局
│   ├── views/               # 路由页面
│   ├── stores/              # Pinia 状态
│   └── router/              # 路由与认证守卫
├── docs/                    # 开发文档
├── scripts/                 # 发布构建、Skill 打包和演示数据
├── skills/rewind/           # Rewind Agent Skill 源文件
├── Dockerfile               # 多架构发布镜像入口
└── .github/workflows/       # 标签发布流程
```

生成目录不应直接编辑：

- `frontend/dist/`：前端构建产物。
- `backend/internal/server/web/dist/`：嵌入后端时的临时资源目录。
- `build/` 或发布脚本指定目录：发布二进制产物。

## 本地开发

从仓库根目录启动后端：

```bash
cd backend
REWIND_ADMIN_PASSWORD=123456 go run ./cmd/rewind serve \
  --listen 127.0.0.1:8080 \
  --data-dir ./data
```

另一个终端启动前端：

```bash
cd frontend
npm run dev -- --host 127.0.0.1
```

Vite 开发服务器默认监听 `http://localhost:5173`，并代理 `/api` 到 `http://localhost:8080`。

可选地导入演示数据：

```bash
REWIND_ADMIN_PASSWORD=123456 python3 scripts/seed-demo.py
```

脚本会创建演示项目、文件和版本；发现同名演示项目时会停止，不修改已有数据。

## 验证命令

```bash
cd backend && go test ./...
cd frontend && npm run type-check
cd frontend && npm run lint
cd frontend && npm run build
```

除非明确要求，不要创建新的单元测试。修改前端时优先运行受影响页面对应的类型检查、lint 或构建；修改后端时运行覆盖受影响包的 `go test`。

完整发布构建：

```bash
./scripts/build.sh
```

脚本会构建前端、临时嵌入资源，并为多个操作系统和架构交叉编译二进制文件。

## 架构约束

### API 和认证

- REST API 统一使用 `/api/v1`。
- 所有 JSON 响应使用 HTTP 200；通过 `success` 判断成功，并读取 `data` 或 `error.code`。
- 管理后台接口使用登录 Token；备份上传只接受 API Token；历史查询和下载可接受 API Token 或登录 Token。
- 登录和管理接口不得使用 API Token，上传接口不得使用登录 Token。
- 下载成功直接输出二进制内容；下载失败才使用统一 JSON 错误响应。
- 实体 ID 在请求中忽略大小写。

### 数据与存储

- 项目、文件和版本元数据保存在 SQLite 中。
- 每次备份生成独立版本，即使文件内容相同也不去重。
- 文件通过 `(project_id, key)` 定位；版本通过版本 ID 定位。
- 存储对象独立保存，并通过 `storage.Store` 接口读取、写入、检查和删除。
- 上传应使用流式处理和暂存文件，并在元数据提交失败时清理已提交的版本文件。
- 删除项目时必须连带清理文件、版本元数据和版本内容。
- 不实现基于数量或时间的自动保留策略。

### 前端

- API 调用集中放在 `frontend/src/api/`，共享响应类型放在 `frontend/src/api/types.ts`。
- 语义或接口契约不同的 API 类型即使字段相同，也应独立声明；不要通过 `extends`、`Pick` 或 `Omit` 从另一 API 类型派生。真正通用的结构（如分页参数）可组合复用。
- 页面和路由状态遵循现有 Pinia store、Vue Router 和 Element Plus 用法。
- 管理页面受路由认证守卫保护；登录和过期状态由 `stores/auth.ts` 管理。

## 配置与安全

- 首次初始化管理员密码只能由 `REWIND_ADMIN_PASSWORD` 或 `REWIND_ADMIN_PASSWORD_FILE` 提供，不能在命令行参数中传递。
- 数据目录默认是当前工作目录下的 `data`；Docker 使用 `/data`。
- Docker 镜像以非 root 用户运行。
- 登录密码只保存安全哈希；登录 Token 只保存哈希。API Token 保存校验哈希及用于管理后台再次查看的明文。
- 日志不得记录密码、完整登录 Token 或完整 API Token。
- API Token 明文存储在 SQLite 中，数据库文件及其备份都应按凭证妥善保护。已有数据库升级后，历史 Token 仍可使用，但因只存有哈希而无法查看。
- 不要提交真实密码、Token、`backend/data/` 数据库或用户备份文件。
