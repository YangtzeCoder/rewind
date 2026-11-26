---
name: rewind
description: 根据用户请求或指令，将本地文件备份到 Rewind 服务、查看版本历史或恢复指定版本。
---

# Rewind

当用户要求使用 Rewind 备份、查看历史或恢复文件，或者用户的指令要求执行这些操作时，使用本 Skill。本 Skill 不决定何时需要备份。运行 `python3 <skill-directory>/scripts/rewind.py --help` 查看命令用法；路径中的 `<skill-directory>` 指已安装的 Skill 目录，不是当前项目目录。

## 配置

确定当前项目根目录，并将其作为工作目录；也可以在子命令前指定 `--project-root ROOT`。全局配置位于 `~/.rewind/config.json`，项目配置位于项目根目录的 `.rewind/config.json`。按字段合并配置：项目配置中的值优先，缺失字段从全局配置继承。

首次配置时，询问用户选择以下方式，默认选择 1：

1. **本地项目 + 全局凭证（默认）**：每个项目本地保存自己的 project ID，所有项目共用全局 URL 和 Token。
2. **全部本地**：将 URL、Token 和 project ID 都保存在当前项目。
3. **全部全局**：URL、Token 和默认 project ID 都保存在全局；未设置本地 project ID 的项目共用此默认项目。

如果本地项目配置不存在，并且配置错误只指出缺少 project_id（说明 URL 和 Token 已可用），不要再展示上述选择，也不要询问 URL/Token；只提示用户提供当前项目的 project ID，并执行 `python3 <skill-directory>/scripts/rewind.py --project-root ROOT configure --project PROJECT_ID`，将其保存到本地。

引导用户在 Rewind 管理后台创建项目和 API Token。Token 可访问所有项目的 Agent 接口。对于需要配置 Token 的方式，使用当前 Agent 平台提供的安全密钥输入能力，让用户输入 Token；由 Agent 执行以下命令之一，并通过安全的非交互标准输入流传入 Token：

- 方式 1：`python3 <skill-directory>/scripts/rewind.py --project-root ROOT configure --scope shared --project PROJECT_ID --url URL --token-stdin`
- 方式 2：`python3 <skill-directory>/scripts/rewind.py --project-root ROOT configure --scope local --project PROJECT_ID --url URL --token-stdin`
- 方式 3：`python3 <skill-directory>/scripts/rewind.py --project-root ROOT configure --scope global --project PROJECT_ID --url URL --token-stdin`

安全输入能力必须确保 Token 不出现在普通聊天、命令参数、终端输出或工具日志中；不要假设普通聊天、工具调用参数或终端输入是安全的。若平台不能同时提供安全输入和安全传输，向用户说明限制并停止配置，不要要求用户运行命令或声称配置成功。

配置写入成功后继续用户原先请求的操作。不要读取或输出配置文件的内容；配置文件包含密钥，不要提交、分享或备份。脚本会为项目配置添加本地 Git 忽略规则，但要提醒用户：忽略规则不会取消 Git 对已跟踪文件的跟踪。

本地配置文件存在但内容无效时，明确报告配置错误，不要静默退回全局配置。仅完成配置不能验证项目和 Token 是否有效，只有实际 API 操作成功后才能报告验证成功。如果 Token 缺失或无效，按上述安全输入流程引导用户设置或替换，不要暴露其值。

## 操作

- 确定用户的项目根目录。项目内文件默认使用相对于根目录的路径作为 key；项目外文件或根目录不明确时，向用户询问稳定的 key，并通过 `--key KEY` 指定。key 区分大小写，不要擅自编造。
- 使用 `backup` 备份用户指定的文件，可一次传入多个文件；不会自动递归备份目录。仅在有用时提供可选描述。逐项报告结果，包括失败项和版本 ID。如果用户要求在编辑前备份而备份失败，应报告失败，并遵循用户关于是否继续的指令。
- 使用 `history --key KEY` 展示版本 ID、时间、描述和大小。根据需要获取后续页面。第一页的第一条是最新版本。
- 使用 `download --version VERSION_ID --output PATH` 恢复文件或另存副本。如果目标路径已存在，先向用户说明选定的版本和路径，取得确认后才能使用 `--force`。版本不明确时先询问用户。下载失败时不得声称文件已恢复。

辅助脚本向标准输出写入 JSON 结果，向标准错误输出错误信息；任何操作失败时返回非零退出码。不要展示、记录凭证，也不要将凭证放入终端命令中。
