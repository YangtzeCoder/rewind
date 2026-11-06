# 项目简介

Rewind 是面向 AI Agent 的自托管文件备份与版本管理服务。项目目标、功能范围和技术方案以 `DOCS/PRD.md` 为准。

## 主要文档与目录

- [`DOCS/PRD.md`](DOCS/PRD.md)：产品需求文档，说明项目管理、文件备份与版本管理、认证、REST API、存储和部署要求。
- [`DOCS/API_DESIGN.md`](DOCS/API_DESIGN.md)：API 设计规范，说明请求、统一响应结构和文件下载约定；错误码以代码为准。
- [`backend/internal/server/`](backend/internal/server/)：后端 HTTP 服务实现目录，负责服务启动和 API 路由。

## 开发规则

- 除非明确要求，否则不要创建单元测试。
