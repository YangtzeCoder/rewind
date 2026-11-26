<script setup lang="ts">
import { ref } from 'vue'

const brandMark = '/brand-mark.svg'
const origin = window.location.origin
const isHttp = window.location.protocol === 'http:'
const installMethods = [
  {
    id: 'github',
    title: '从 GitHub 安装',
    text: 'npx skills add https://github.com/YangtzeCoder/rewind',
    copyLabel: '复制 GitHub 安装命令',
    successMessage: 'GitHub 安装命令已复制。',
  },
  {
    id: 'service',
    title: '从当前服务安装',
    text: `npx skills add ${origin}`,
    copyLabel: '复制当前服务安装命令',
    successMessage: '当前服务安装命令已复制。',
  },
  {
    id: 'agent',
    title: '交给 Agent 安装',
    text: `请根据 ${origin}/.well-known/skills/index.json 中的说明，获取并安装 Rewind Skill`,
    copyLabel: '复制 Agent 安装提示词',
    successMessage: 'Agent 安装提示词已复制。',
  },
]
const copyError = '复制失败，请手动选择并复制内容。'
const copyStatus = ref<{ id: string; succeeded: boolean; message: string } | null>(null)

async function copyMethod(method: (typeof installMethods)[number]) {
  try {
    await navigator.clipboard.writeText(method.text)
    copyStatus.value = { id: method.id, succeeded: true, message: method.successMessage }
  } catch {
    copyStatus.value = { id: method.id, succeeded: false, message: copyError }
  }
}
</script>

<template>
  <div class="home">
    <header class="home-header">
      <div class="home-brand">
        <img :src="brandMark" alt="" width="32" height="32" />
        <span>Rewind</span>
      </div>
      <nav class="home-nav" aria-label="首页导航">
        <a href="https://github.com/YangtzeCoder/rewind" target="_blank" rel="noopener noreferrer">
          GitHub
        </a>
        <RouterLink to="/login">登录</RouterLink>
      </nav>
    </header>

    <main>
      <section class="hero" aria-labelledby="hero-title">
        <h1 id="hero-title">
          <span>为 AI Agent</span>
          <span>保留文件版本</span>
        </h1>
        <p>开源、自托管的文件备份服务，让 Agent 保存重要文件，需要时找回历史版本。</p>
        <a class="install-link" href="#install">安装 Skill</a>
      </section>

      <section id="install" class="install" aria-labelledby="install-title" tabindex="-1">
        <div class="install-heading">
          <h2 id="install-title">安装 Rewind Skill</h2>
          <p>使用 npx 安装需要 Node.js 22.20 或更新版本，以及 Python 3。</p>
        </div>

        <div v-for="method in installMethods" :key="method.id" class="install-method">
          <h3>{{ method.title }}</h3>
          <div class="method-content">
            <div class="command-box">
              <code v-if="method.id !== 'agent'">{{ method.text }}</code>
              <p v-else class="agent-prompt">{{ method.text }}</p>
              <button type="button" :aria-label="method.copyLabel" @click="copyMethod(method)">
                {{ copyStatus?.id === method.id && copyStatus.succeeded ? '已复制' : '复制' }}
              </button>
            </div>
            <p v-if="method.id === 'service'" class="method-note">
              <template v-if="isHttp">
                当前页面使用 HTTP。请访问指向此 Rewind 服务的 HTTPS 地址后安装。
              </template>
              <template v-else>安装地址须使用 HTTPS，并指向此 Rewind 服务。</template>
            </p>
            <p v-if="copyStatus?.id === method.id && !copyStatus.succeeded" class="copy-error">
              {{ copyStatus.message }}
            </p>
          </div>
        </div>

        <span class="copy-status" role="status">{{ copyStatus?.message }}</span>
      </section>
    </main>
  </div>
</template>

<style scoped>
.home {
  --home-background: #fff;
  --home-text: #192b41;
  --home-muted: #486078;
  --home-accent: #1d6fbe;
  --home-accent-hover: #175c9f;
  --home-divider: #e9eff6;
  --home-command: #f4f8fc;
  --home-gutter: clamp(20px, 5vw, 64px);

  min-height: 100vh;
  padding-inline: var(--home-gutter);
  background: var(--home-background);
  color: var(--home-text);
}

.home-header {
  max-width: 1120px;
  min-height: 76px;
  margin-inline: auto;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  border-bottom: 1px solid var(--home-divider);
}

.home-brand {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 21px;
  font-weight: 700;
  letter-spacing: -0.04em;
}

.home-brand img {
  display: block;
}

.home-nav {
  display: flex;
  align-items: center;
  gap: clamp(16px, 3vw, 32px);
}

.home-nav a {
  display: inline-flex;
  align-items: center;
  min-height: 44px;
  color: var(--home-muted);
  font-size: 14px;
  font-weight: 600;
  text-decoration: none;
}

.home-nav a:hover {
  color: var(--home-accent);
  text-decoration: underline;
  text-underline-offset: 4px;
}

.hero {
  padding-block: clamp(72px, 9vw, 136px) clamp(72px, 8vw, 112px);
  text-align: center;
}

.hero h1 {
  margin: 0;
  font-size: clamp(36px, 5.4vw, 76px);
  font-weight: 750;
  line-height: 1.14;
  letter-spacing: -0.04em;
}

.hero h1 span {
  display: inline-block;
}

.hero h1 span + span {
  margin-inline-start: 0.14em;
}

.hero p {
  max-width: 680px;
  margin: 24px auto 0;
  color: var(--home-muted);
  font-size: clamp(16px, 1.4vw, 18px);
  line-height: 1.85;
  text-wrap: balance;
}

.install-link {
  display: inline-flex;
  align-items: center;
  min-height: 48px;
  margin-top: 30px;
  padding-inline: 24px;
  border-radius: 8px;
  background: var(--home-accent);
  color: white;
  font-size: 15px;
  font-weight: 600;
  text-decoration: none;
}

.install-link:hover {
  background: var(--home-accent-hover);
}

.home a:focus-visible,
.home button:focus-visible {
  outline: 2px solid var(--home-accent);
  outline-offset: 3px;
}

.install {
  max-width: 880px;
  margin-inline: auto;
  padding-block: clamp(36px, 5vw, 56px) clamp(56px, 7vw, 96px);
  border-top: 1px solid var(--home-divider);
  scroll-margin-top: 24px;
}

.install-heading {
  margin-bottom: 24px;
}

.install-heading h2 {
  margin: 0;
  font-size: clamp(26px, 3vw, 36px);
  font-weight: 700;
  line-height: 1.25;
  letter-spacing: -0.035em;
}

.install-heading p,
.method-note,
.copy-error {
  font-size: 14px;
  line-height: 1.65;
}

.install-heading p {
  margin: 10px 0 0;
  color: var(--home-muted);
}

.install-method {
  display: grid;
  grid-template-columns: 150px minmax(0, 1fr);
  align-items: start;
  gap: 20px;
  padding-block: 16px;
  border-top: 1px solid var(--home-divider);
}

.install-method h3 {
  margin: 15px 0 0;
  font-size: 15px;
  font-weight: 650;
  line-height: 1.6;
}

.command-box {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 5px 6px 5px 14px;
  border-radius: 8px;
  background: var(--home-command);
}

.command-box code {
  font-family: ui-monospace, SFMono-Regular, Consolas, monospace;
}

.command-box code,
.agent-prompt {
  min-width: 0;
  margin: 0;
  padding-block: 5px;
  font-size: 13px;
  line-height: 1.6;
  overflow-wrap: anywhere;
}

.command-box button {
  flex: none;
  min-width: 64px;
  min-height: 44px;
  padding: 0 8px;
  border: 1px solid var(--home-divider);
  border-radius: 6px;
  background: var(--home-background);
  color: var(--home-accent);
  font: inherit;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
}

.command-box button:hover {
  border-color: var(--home-accent);
}

.method-note,
.copy-error {
  margin: 8px 0 0;
}

.method-note {
  color: var(--home-muted);
}

.copy-error {
  color: #9a3425;
}

.copy-status {
  position: absolute;
  width: 1px;
  height: 1px;
  overflow: hidden;
  clip-path: inset(50%);
  white-space: nowrap;
}

@media (max-width: 600px) {
  .home-header {
    min-height: 64px;
  }

  .hero h1 span {
    display: block;
  }

  .hero h1 span + span {
    margin-inline-start: 0;
  }

  .install-method {
    grid-template-columns: minmax(0, 1fr);
    gap: 8px;
  }

  .install-method h3 {
    margin-top: 0;
  }
}
</style>
