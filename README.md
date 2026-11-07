# Rewind

Rewind 是面向 AI Agent 的自托管文件备份与版本管理服务，帮助用户管理项目文件、保存备份并查看历史版本。

## Docker 部署

从 GitHub Container Registry 拉取镜像并启动服务。首次启动前设置管理员初始密码；镜像默认通过 `--data-dir /data` 将数据保存在 Docker Volume 中。自定义数据目录时，可在镜像名后指定 `serve --data-dir <目录>` 并挂载对应目录。

```bash
export REWIND_ADMIN_PASSWORD='替换为强密码'
docker run -d --name rewind \
  -p 8080:8080 \
  -v rewind-data:/data \
  -e REWIND_ADMIN_PASSWORD \
  ghcr.io/yangtzecoder/rewind:latest
```

发布版本也可使用对应版本标签，例如 `ghcr.io/yangtzecoder/rewind:v1.2.3`。首次启动后，管理员密码保存在数据目录中，之后更改环境变量不会覆盖已有密码。

## 安装 Agent Skill

需要 Node.js（`npx skills` 要求 Node.js 22.20 或更新版本）以及运行 Skill 辅助脚本所需的 Python 3。从项目目录执行以下命令，按提示选择需要安装的 Agent：

```bash
npx skills add https://github.com/YangtzeCoder/rewwind
```

添加 `-g` 可安装到用户级目录。也可以使用已部署 Skill 分发资源的 Rewind 服务地址安装；该地址的 HTTPS 和流量须指向该服务。

## 配置和使用

1. 在 Rewind 管理后台创建项目和 API Token，记下项目 ID。Token 明文仅在创建时显示一次，且可访问所有项目的 Agent 接口。
2. 告诉 Agent 使用 Rewind。首次使用时，Agent 会引导你在自己的终端运行已安装 Skill 中的脚本：

   ```bash
   python3 <已安装的 Skill 路径>/scripts/rewind.py --project-root <项目根目录> configure --project <项目ID> --url <Rewind服务地址>
   ```

   服务地址由用户提供，例如自己的 Rewind 部署地址。按提示输入 Token，输入内容不会显示。不要把 Token 粘贴到聊天或放进命令参数。
3. 配置保存在该项目根目录的 `.rewind/config.json`，包含 Token；不同项目分别配置。脚本会在 `.rewind/.gitignore` 中添加忽略规则，但这不能取消 Git 对已跟踪文件的跟踪，提交前请确认配置文件不会被提交或分享。
4. 对 Agent 说“备份 `docs/design.md`”“查看 `docs/design.md` 的备份历史”，或“将它恢复到指定版本”。覆盖本地文件前，Agent 会说明版本和路径并请求确认。

用户可以在所用 Agent 支持的项目指令文件中规定备份时机，例如“修改需求文档前先使用 Rewind 备份”。Skill 本身不设置自动备份规则；Agent 是否自动执行取决于它是否读取并遵循相应指令。默认以项目内相对路径作为文件 key；目录不会自动递归备份。

## 开源协议

本项目采用 [MIT 协议](LICENSE)。
