<script setup lang="ts">
import { ref } from 'vue'

const githubCommand = 'npx skills add https://github.com/YangtzeCoder/rewind'
const serviceCommand = `npx skills add ${window.location.origin}`
const isHttp = window.location.protocol === 'http:'
const copyStatus = ref('')

async function copyCommand(command: string, source: string) {
  try {
    await navigator.clipboard.writeText(command)
    copyStatus.value = `${source}命令已复制。`
  } catch {
    copyStatus.value = '复制失败，请手动选择并复制命令。'
  }
}
</script>

<template>
  <main class="home">
    <div class="home-wrap">
      <header class="home-header">
        <div class="home-brand" aria-label="Rewind">
          <span class="home-brand-mark">R<span>↶</span></span>
          <span>rewind<span class="brand-period">.</span></span>
        </div>
        <nav class="home-nav" aria-label="首页导航">
          <a class="header-link" href="#install">安装 Skill <span aria-hidden="true">↗</span></a>
          <a class="header-link" href="https://github.com/YangtzeCoder/rewind" target="_blank" rel="noopener noreferrer">GitHub <span aria-hidden="true">↗</span></a>
        </nav>
      </header>

      <section class="hero" aria-labelledby="hero-title">
        <div class="hero-copy">
          <p class="home-eyebrow"><span></span> 为 AI AGENT 而构建</p>
          <h1 id="hero-title">每一次改变，<br /><span>都有回路。</span></h1>
          <p class="hero-description">Rewind 是面向 AI Agent 的自托管文件备份与版本管理服务。为重要文件留下记录，让你随时找回需要的版本。</p>
          <a class="primary-link" href="#install">开始使用 <span aria-hidden="true">↗</span></a>
          <p class="hero-note">开源 · 自托管 · 数据由你掌控</p>
        </div>

        <div class="hero-art" aria-label="文件版本历史示意图">
          <div class="art-top"><span>YOUR WORK, PRESERVED</span><span>↶</span></div>
          <div class="art-orbit art-orbit-one"></div><div class="art-orbit art-orbit-two"></div>
          <div class="version-card">
            <div class="version-head"><div><span class="file-icon">▤</span><span><strong>design.md</strong><small>项目文件</small></span></div><span class="version-menu">···</span></div>
            <div class="version-label">版本历史 <span>示意</span></div>
            <div class="version-row"><span class="version-dot current"></span><div><strong>最新版本</strong><small>保留当前的想法</small></div><span>↗</span></div>
            <div class="version-row"><span class="version-dot"></span><div><strong>之前的版本</strong><small>需要时随时找回</small></div><span>↶</span></div>
          </div>
          <div class="art-caption"><span>KEEP THE<br />MOMENTS THAT<br /><em>MATTER.</em></span><small>01 / 你的工作，值得被保存</small></div>
        </div>
      </section>

      <section class="features" aria-labelledby="features-title">
        <div class="section-heading"><p class="home-eyebrow">BUILT FOR YOUR WORKFLOW</p><h2 id="features-title">让每一步，都有迹可循。</h2><p>从备份到找回，Rewind 为你和 Agent 的协作保留一份可靠的记录。</p></div>
        <div class="feature-grid">
          <article><span class="feature-number">01 /</span><div class="feature-icon" aria-hidden="true">⌂</div><h3>你的数据，你做主</h3><p>自行部署 Rewind，备份文件和版本数据保存在你管理的服务中。</p></article>
          <article><span class="feature-number">02 /</span><div class="feature-icon" aria-hidden="true">↗</div><h3>为文件留个备份</h3><p>通过 Skill 让 Agent 根据你的请求或指令，保存项目中的重要文件。</p></article>
          <article><span class="feature-number">03 /</span><div class="feature-icon" aria-hidden="true">↶</div><h3>回到需要的版本</h3><p>查看文件的历史记录，下载或恢复你需要的版本。</p></article>
        </div>
      </section>

      <section id="install" class="install" aria-labelledby="install-title">
        <div class="install-intro"><p class="home-eyebrow">GET STARTED / AGENT SKILL</p><h2 id="install-title">让你的 Agent<br /><span>也会使用 Rewind。</span></h2><p>在项目目录中运行下面任意一条命令，按提示选择需要安装的 Agent。</p><div class="requirements"><span>安装前准备</span><p>Node.js 22.20 或更新版本，以及 Python 3。</p></div></div>
        <div class="install-commands">
          <article class="command-card"><div class="command-heading"><div><span>01 / GITHUB</span><h3>从 GitHub 安装</h3></div><span aria-hidden="true">↗</span></div><div class="command-box"><code>{{ githubCommand }}</code><button type="button" aria-label="复制 GitHub 安装命令" @click="copyCommand(githubCommand, 'GitHub 安装')">复制</button></div></article>
          <article class="command-card"><div class="command-heading"><div><span>02 / THIS INSTANCE</span><h3>从当前服务安装</h3></div><span aria-hidden="true">↗</span></div><div class="command-box"><code>{{ serviceCommand }}</code><button type="button" aria-label="复制当前服务安装命令" @click="copyCommand(serviceCommand, '当前服务安装')">复制</button></div><p class="command-help">自托管安装需要 HTTPS，且该地址的流量须指向此 Rewind 服务。<template v-if="isHttp">当前页面使用 HTTP，请通过该服务的 HTTPS 地址访问后安装。</template></p></article>
          <p class="copy-status" role="status" aria-live="polite">{{ copyStatus }}</p>
          <p class="install-note">想安装到用户目录？在命令末尾添加 <code>-g</code>。</p>
        </div>
      </section>

      <footer class="home-footer"><div class="home-brand footer-brand"><span class="home-brand-mark">R<span>↶</span></span><span>rewind<span class="brand-period">.</span></span></div><p>为你的工作留一份安心。</p><span>SELF-HOSTED · OPEN SOURCE</span></footer>
    </div>
  </main>
</template>

<style scoped>
.home{background:#fff;color:#192b41;min-height:100vh}.home-wrap{max-width:1440px;margin:auto;padding:0 7%}.home-header{height:96px;display:flex;align-items:center;justify-content:space-between;border-bottom:1px solid #e9eff6}.home-brand{display:flex;align-items:center;gap:10px;font-size:23px;font-weight:750;letter-spacing:-1px}.home-brand-mark{width:36px;height:36px;border-radius:9px;display:grid;place-items:center;position:relative;background:#409eff;color:#fff;font-size:21px;letter-spacing:0}.home-brand-mark span{position:absolute;right:3px;bottom:2px;font-size:11px}.brand-period{color:#409eff}.home-nav{display:flex;align-items:center;gap:24px}.header-link{font-size:13px;font-weight:600;color:#486078;text-decoration:none}.header-link:hover{color:#2586e8}.header-link span{margin-left:8px;color:#409eff}.hero{display:grid;grid-template-columns:1fr 1fr;gap:7%;align-items:center;padding:68px 0 90px}.hero-copy{padding:20px 0}.home-eyebrow{font-size:10px;font-weight:700;letter-spacing:2px;color:#5283b7;margin:0 0 24px}.hero .home-eyebrow span{display:inline-block;width:7px;height:7px;border-radius:50%;background:#409eff;margin-right:7px}.hero h1,.install h2{font-size:clamp(38px,4.4vw,65px);letter-spacing:-2px;line-height:1.3;margin:0;font-weight:700}.hero h1 span,.install h2 span{color:#409eff}.hero-description{font-size:16px;line-height:2;color:#718399;max-width:470px;margin:25px 0 32px}.primary-link{display:inline-flex;align-items:center;gap:38px;background:#409eff;color:white;text-decoration:none;font-size:14px;font-weight:600;padding:15px 20px;border-radius:6px}.primary-link:hover{background:#2586e8}.hero-note{font-size:12px;color:#9aaabd;margin:22px 0 0}.hero-art{position:relative;overflow:hidden;min-height:470px;border-radius:10px;background:#eaf4ff;padding:26px;color:#4d8ed2}.art-top{display:flex;justify-content:space-between;font-size:9px;letter-spacing:1.6px;font-weight:700}.art-top span:last-child{font-size:24px;line-height:1}.art-orbit{position:absolute;border:1px solid #c8e2fc;border-radius:50%;width:390px;height:390px;top:65px;right:-70px}.art-orbit-two{width:490px;height:490px;top:15px;right:-120px}.version-card{position:relative;z-index:1;width:min(82%,350px);margin:54px auto 0;background:#fff;box-shadow:0 18px 55px #8ab5dd30;border:1px solid #e4effa;border-radius:10px;padding:20px;color:#273c54}.version-head,.version-head>div{display:flex;align-items:center}.version-head{justify-content:space-between}.version-head>div{gap:10px}.file-icon{display:grid;place-items:center;width:38px;height:38px;border-radius:7px;color:#409eff;background:#eaf4ff;font-size:21px}.version-head strong,.version-head small,.version-row strong,.version-row small{display:block}.version-head strong,.version-row strong{font-size:12px}.version-head small,.version-row small{font-size:10px;color:#9aaabd;margin-top:5px}.version-menu{color:#a5b5c5;font-size:18px}.version-label{border-top:1px solid #edf2f7;margin-top:20px;padding-top:17px;font-size:11px;font-weight:700}.version-label span{color:#a6b5c4;font-size:9px;font-weight:400;margin-left:4px}.version-row{display:flex;align-items:center;gap:12px;padding-top:16px}.version-row>div{flex:1}.version-row>span:last-child{color:#89a4bf}.version-dot{width:9px;height:9px;border:2px solid #b5d2ee;border-radius:50%}.version-dot.current{background:#409eff;border-color:#409eff}.art-caption{position:absolute;bottom:25px;left:26px;right:26px;display:flex;align-items:end;justify-content:space-between;font-size:13px;font-weight:700;line-height:1.25;letter-spacing:1px}.art-caption em{font-family:Georgia,serif;font-weight:400}.art-caption small{font-size:9px;letter-spacing:1px;font-weight:400}.features{padding:20px 0 110px}.section-heading{max-width:620px}.section-heading h2{font-size:clamp(27px,3vw,38px);letter-spacing:-1px;margin:0 0 14px}.section-heading>p:last-child{color:#8494a6;line-height:1.8;font-size:14px}.feature-grid{display:grid;grid-template-columns:repeat(3,1fr);gap:0;margin-top:46px;border-top:1px solid #e8eff6;border-bottom:1px solid #e8eff6}.feature-grid article{position:relative;padding:30px 30px 32px 0}.feature-grid article+article{padding-left:30px;border-left:1px solid #e8eff6}.feature-number{font-size:10px;color:#a8b7c8;letter-spacing:1px}.feature-icon{font-size:29px;color:#409eff;margin:24px 0 14px;line-height:1}.feature-grid h3{font-size:16px;margin:0 0 10px}.feature-grid p{font-size:13px;line-height:1.9;color:#8191a3;margin:0;max-width:280px}.install{background:#f2f8ff;border-radius:10px;padding:60px;display:grid;grid-template-columns:0.9fr 1.1fr;gap:9%;scroll-margin-top:20px}.install h2{font-size:clamp(28px,3vw,40px);letter-spacing:-1px}.install-intro>p:not(.home-eyebrow){font-size:14px;line-height:1.9;color:#71879d;margin:20px 0;max-width:360px}.requirements{margin-top:45px;border-top:1px solid #d8e9f9;padding-top:18px;font-size:11px;color:#6885a1}.requirements span{font-weight:700;color:#56779a}.requirements p{margin:8px 0;line-height:1.7}.command-card{background:#fff;border:1px solid #e2eef9;border-radius:8px;padding:22px;margin-bottom:12px}.command-heading{display:flex;justify-content:space-between;color:#8aa9c8}.command-heading span{font-size:9px;letter-spacing:1.5px;font-weight:700}.command-heading>span{font-size:17px}.command-heading h3{color:#253e59;font-size:15px;margin:8px 0 16px}.command-box{display:flex;align-items:center;gap:12px;background:#f4f8fc;border-radius:5px;padding:12px}.command-box code{flex:1;min-width:0;color:#35638e;font-size:11px;line-height:1.6;overflow-wrap:anywhere}.command-box button{flex-shrink:0;background:white;border:1px solid #dce9f5;border-radius:4px;padding:6px 9px;color:#4a80b4;font-size:11px;cursor:pointer}.command-box button:hover{border-color:#409eff}.command-box button:focus-visible,.primary-link:focus-visible,.header-link:focus-visible{outline:2px solid #2179cf;outline-offset:3px}.command-help,.install-note{font-size:11px;line-height:1.8;color:#8499ad}.command-help{margin:12px 0 0}.copy-status{min-height:18px;font-size:12px;color:#397cb8;margin:0}.install-note{margin:4px 0 0}.install-note code{color:#397cb8}.home-footer{display:flex;align-items:center;gap:25px;padding:44px 0;color:#a0afbf}.footer-brand{font-size:17px}.footer-brand .home-brand-mark{width:28px;height:28px;font-size:16px}.footer-brand .home-brand-mark span{font-size:9px}.home-footer p{font-size:12px;margin:0;flex:1}.home-footer>span{font-size:9px;letter-spacing:1.4px}
@media(max-width:800px){.home-wrap{padding:0 6%}.home-header{height:76px}.hero{grid-template-columns:1fr;gap:35px;padding:45px 0 65px}.hero h1{font-size:clamp(38px,8vw,58px)}.hero-art{min-height:410px}.features{padding:10px 0 70px}.feature-grid{grid-template-columns:1fr;margin-top:30px}.feature-grid article,.feature-grid article+article{padding:24px 0;border-left:0}.feature-grid article+article{border-top:1px solid #e8eff6}.feature-icon{margin:16px 0 10px}.feature-grid p{max-width:none}.install{grid-template-columns:1fr;gap:28px;padding:30px 24px}.requirements{margin-top:25px}.home-footer{flex-wrap:wrap;gap:14px;padding:32px 0}.home-footer p{flex:1}.home-footer>span{width:100%}}
@media(max-width:380px){.home-nav{gap:12px}.header-link span{margin-left:4px}.hero-art{min-height:380px}.version-card{width:95%;margin-top:42px}.art-caption small{display:none}.command-box{align-items:flex-start}.home-footer p{font-size:11px}}
</style>
