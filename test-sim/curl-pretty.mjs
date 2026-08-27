#!/usr/bin/env node
// ============================================================================
//  curl-pretty — 把 API 的 JSON 响应美化输出，超长字符串（base64 图片等）截断为摘要
//
//  用法（Windows cmd / PowerShell 均可）：
//    curl -s -X POST http://localhost:8080/api/user/signin -H "Content-Type: application/json" -d "{\"email\":\"...\",\"password\":\"...\"}" | node test-sim/curl-pretty.mjs
//    ... | node test-sim/curl-pretty.mjs --save-images   # 额外把 data:image 另存为文件
//
//  不需要 jq，只需项目已有的 Node。
// ============================================================================
import { writeFileSync } from 'fs'

const args = process.argv.slice(2)
const SAVE = args.includes('--save-images')
const MAX = 80

let raw = ''
process.stdin.setEncoding('utf8')
process.stdin.on('data', (d) => (raw += d))
process.stdin.on('end', () => {
  let obj
  try {
    obj = JSON.parse(raw)
  } catch {
    // 不是 JSON（如错误页）就原样输出
    process.stdout.write(raw)
    return
  }

  const saved = []
  const out = transform(obj, saved)

  if (SAVE && saved.length) console.log(`[已保存图片] ${saved.join(', ')}`)
  console.log(JSON.stringify(out, null, 2))
})

function transform(v, saved) {
  if (typeof v === 'string') {
    if (SAVE && v.startsWith('data:image/')) {
      const m = v.match(/^data:image\/([a-zA-Z0-9+.-]+);base64,(.*)$/s)
      if (m) {
        const ext = m[1].replace('jpeg', 'jpg').split('+')[0] || 'img'
        const name = `saved-image-${saved.length + 1}.${ext}`
        try {
          writeFileSync(name, Buffer.from(m[2], 'base64'))
          saved.push(name)
          return `[图片已保存为 ${name}]`
        } catch {
          // 写入失败则继续走截断逻辑
        }
      }
    }
    if (v.length > MAX) {
      return v.slice(0, 40) + `…[${v.length}字符已省略]`
    }
    return v
  }
  if (Array.isArray(v)) return v.map((x) => transform(x, saved))
  if (v && typeof v === 'object') {
    const out = {}
    for (const [k, x] of Object.entries(v)) out[k] = transform(x, saved)
    return out
  }
  return v
}
