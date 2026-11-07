// Package the repository's skill files for the well-known discovery endpoint.
import { cp, mkdir, readdir, writeFile } from 'node:fs/promises'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const source = path.join(root, 'skills', 'rewind')
const output = path.join(root, 'frontend', 'dist', '.well-known', 'skills')

async function filesIn(directory, prefix = '') {
  const result = []
  for (const entry of await readdir(directory, { withFileTypes: true })) {
    const relative = path.posix.join(prefix, entry.name)
    if (entry.isDirectory()) result.push(...await filesIn(path.join(directory, entry.name), relative))
    else if (entry.isFile() && !entry.name.startsWith('.') && !entry.name.endsWith('.pyc')) result.push(relative)
  }
  return result.sort()
}

const files = await filesIn(source)
if (!files.includes('SKILL.md')) throw new Error('Rewind skill is missing SKILL.md')
await mkdir(output, { recursive: true })
await cp(source, path.join(output, 'rewind'), {
  recursive: true,
  filter: (src) => !path.basename(src).startsWith('.') && path.basename(src) !== '__pycache__' && !src.endsWith('.pyc'),
})
await writeFile(path.join(output, 'index.json'), JSON.stringify({
  skills: [{ name: 'rewind', description: '根据用户请求或指令，将本地文件备份到 Rewind 服务、查看版本历史或恢复指定版本。', files }],
}, null, 2) + '\n')
