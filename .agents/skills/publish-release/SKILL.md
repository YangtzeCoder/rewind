---
name: publish-release
description: 为 Rewind 选择版本号、检查发布条件、推送当前分支和版本标签，并验证 GitHub Actions 发布结果。
---

# 发布 Rewind 版本

当用户要求发布 Rewind 新版本时使用本 Skill。Skill 位于当前仓库的 `.agents/skills/publish-release/`，只供此项目使用；不要复制到用户级或全局 Skill 目录。

## 版本选择

1. 检查本地和远程已发布的 `vMAJOR.MINOR.PATCH` 标签，确定最新稳定版本。不要把预发布标签当作稳定版本基准。
2. 以最新稳定版本为基准，分别只将主版本、次版本、修订版本加 1，向用户提供三个候选版本并请用户选择。比如最新版本为 `v0.2.3`，候选为 `v1.2.3`、`v0.3.3`、`v0.2.4`。
3. 检查用户选择的标签在本地和远端都不存在。如果候选已被占用，说明冲突并请用户从可用候选或自定义版本中选择；不要覆盖或重用已发布标签。
4. 如果仓库没有稳定版本标签，或无法确认最新版本，先向用户询问，不要自行猜测。

## 发布步骤

1. 确认当前目录属于 Rewind 仓库，读取当前分支、远程地址、工作区状态及分支与远程的差异。需要发布的应是当前分支的 HEAD。
2. 工作区有未提交修改时暂停并告知用户；不要自动暂存、提交、丢弃或隐藏修改。处于 detached HEAD、没有明确推送远程，或本地分支与远程分支发生分叉时暂停并询问用户。不得强制推送。
3. 发布前只验证前端和后端能否在当前环境成功构建，不运行会交叉编译其他操作系统或 CPU 架构的发布脚本。运行：

   ```bash
   cd frontend && npm ci && npm run build
   cd ../backend && go build ./...
   ```

   前端构建包含类型检查；后端仅针对当前环境编译所有 Go 包。
4. 记录当前分支名、远程名和要发布的提交 SHA。**先推送当前分支，再推送标签**，确保发布提交也可从远程分支访问。不要只推送标签：

   ```bash
   git push <remote> <commit>:refs/heads/<branch>
   git tag -a <tag> <commit> -m "Release <tag>"
   git push <remote> refs/tags/<tag>
   ```

   `<commit>` 是已检查的当前 HEAD SHA，`<branch>` 是当前分支名，`<remote>` 是确认过的远程名，`<tag>` 是用户选定的 `vMAJOR.MINOR.PATCH`。若分支推送失败，停止发布标签并报告错误；不要强推或绕过错误。
5. 确认远程分支和标签都指向预期提交 SHA。推送 `v*` 标签会触发 `.github/workflows/release.yml`。
6. 等待并检查 GitHub Actions Release 工作流。只有工作流成功且 GitHub Release 已创建，才能报告发布完成；失败时查看失败作业日志并如实报告，不能将“标签推送成功”说成“版本发布成功”。
7. 成功后提供 GitHub Release、工作流运行记录和镜像标签，并明确报告推送的分支、标签及提交 SHA。如果工作流仍在运行，说明当前状态，不要声称已完成。

发布工作流会生成多平台二进制文件和 SHA256 校验文件，构建 `linux/amd64`、`linux/arm64` 镜像并发布到 GHCR。只有稳定版本会更新 `latest`。
