// ============================================================================
//  POST-FIX VERIFICATION — confirm every reported finding is now closed
// ============================================================================
const BASE = 'http://localhost:8080/api/'
const uniq = (p) => `${p}_${Date.now()}_${Math.floor(Math.random() * 1e6)}`
let closed = 0, open = 0
const notes = []
function check(name, ok, ev) {
  if (ok) { closed++; console.log(`  🟢 CLOSED ${name}${ev ? '  -> ' + ev : ''}`) }
  else { open++; console.log(`  🔴 STILL OPEN ${name}${ev ? '  -> ' + ev : ''}`) }
}
async function req(method, url, { token, body } = {}) {
  const headers = { 'content-type': 'application/json' }
  if (token) headers.authorization = `Bearer ${token}`
  const res = await fetch(BASE + url, { method, headers, body: body ? JSON.stringify(body) : undefined })
  const ct = res.headers.get('content-type') || ''
  const data = ct.includes('json') ? await res.json() : await res.text()
  return { status: res.status, data }
}

console.log('===== SETUP =====')
const email = uniq('fix'), email2 = uniq('fix2')
const su = (await req('POST', 'user/signup', { body: { email, password: 'victim123', firstName: 'Fix', lastName: 'User' } })).data
const token = su.token, uid = su.result._id
const su2 = (await req('POST', 'user/signup', { body: { email: email2, password: 'other123', firstName: 'Other' } })).data
const token2 = su2.token, uid2 = su2.result._id

console.log('\n===== P0-1: getUser 鉴权 + 密码剔除 =====')
const anon = await req('GET', `user/getUser/${uid}`)
check('getUser without token rejected (401)', anon.status === 401, `status=${anon.status}`)
const authed = await req('GET', `user/getUser/${uid2}`, { token })
check('getUser with token works (200)', authed.status === 200, `status=${authed.status}`)
check('getUser response has NO password (empty/absent)', !(authed.data?.user?.password), `password=${JSON.stringify(authed.data?.user?.password)}`)
const si = await req('POST', 'user/signin', { body: { email, password: 'victim123' } })
check('signin response has NO password (empty/absent)', !(si.data?.result?.password), `password=${JSON.stringify(si.data?.result?.password)}`)
const search = await req('GET', `posts/search?searchQuery=${encodeURIComponent('Fix')}`, { token })
const sugUsers = search.data?.user || []
check('search returns users without password', sugUsers.every((u) => !u.password), `users=${sugUsers.length}`)

console.log('\n===== P0-2: 通知接口鉴权 =====')
const notifAnon = await req('GET', `notification/${uid}`)
check('notification read without token rejected', notifAnon.status === 401, `status=${notifAnon.status}`)
const notifOther = await req('GET', `notification/${uid2}`, { token })
check('notification cross-user read rejected', notifOther.status === 401, `status=${notifOther.status}`)
const markOther = await req('PATCH', `notification/mark-notification-as-readed/${uid2}`, { token })
check('notification cross-user mark rejected', markOther.status === 401, `status=${markOther.status}`)
const notifOwn = await req('GET', `notification/${uid}`, { token })
check('notification own read works', notifOwn.status === 200, `status=${notifOwn.status}`)

console.log('\n===== P0-3: WebSocket 鉴权 =====')
const wsNoToken = await new Promise((resolve) => {
  let ws; try { ws = new WebSocket(`ws://localhost:8088/ws/${uid}`) } catch { return resolve('ERR') }
  let opened = false
  ws.onopen = () => { opened = true; setTimeout(() => { try { ws.close() } catch { } resolve('OPENED') }, 1500) }
  ws.onerror = () => resolve('REJECTED')
  ws.onclose = () => resolve(opened ? 'OPENED' : 'REJECTED')
  setTimeout(() => { try { ws.close() } catch { } resolve(opened ? 'OPENED' : 'REJECTED') }, 4000)
})
check('notify WS without token rejected', wsNoToken === 'REJECTED' || wsNoToken === 'ERR', `result=${wsNoToken}`)
const wsWithToken = await new Promise((resolve) => {
  let ws; try { ws = new WebSocket(`ws://localhost:8088/ws/${uid}?token=${encodeURIComponent(token)}`) } catch { return resolve('ERR') }
  let opened = false
  ws.onopen = () => { opened = true; setTimeout(() => { try { ws.close() } catch { } resolve('OPENED') }, 1500) }
  ws.onerror = () => resolve('REJECTED')
  setTimeout(() => { try { ws.close() } catch { } resolve(opened ? 'OPENED' : 'REJECTED') }, 4000)
})
check('notify WS with valid token opens', wsWithToken === 'OPENED', `result=${wsWithToken}`)
const chatNoToken = await new Promise((resolve) => {
  let ws; try { ws = new WebSocket(`ws://localhost:8001/ws/${uid}`) } catch { return resolve('ERR') }
  let opened = false
  ws.onopen = () => { opened = true; setTimeout(() => { try { ws.close() } catch { } resolve('OPENED') }, 1500) }
  ws.onerror = () => resolve('REJECTED')
  setTimeout(() => { try { ws.close() } catch { } resolve(opened ? 'OPENED' : 'REJECTED') }, 4000)
})
check('chat WS without token rejected', chatNoToken === 'REJECTED' || chatNoToken === 'ERR', `result=${chatNoToken}`)

console.log('\n===== P1: 频率限制 / 登录锁定 =====')
// 发帖限流：连续发帖直到触发 429（上限 30/分钟）
let post429 = false, createdIds = []
for (let i = 0; i < 32; i++) {
  const r = await req('POST', 'posts', { token, body: { title: `rl${i}`, message: 'x' } })
  if (r.status === 429) { post429 = true; break }
  if (r.data?._id) createdIds.push(r.data._id)
}
check('post creation rate limit triggers 429', post429, `429 after ${createdIds.length} posts`)
for (const id of createdIds) await req('DELETE', `posts/${id}`, { token })

// 登录锁定：连续 6 次错密码 → 第 6 次应 429
let lockSeen = false
for (let i = 0; i < 6; i++) {
  const r = await req('POST', 'user/signin', { body: { email, password: `wrong${i}` } })
  if (r.status === 429) { lockSeen = true; break }
}
check('login lockout triggers 429 after 5 failures', lockSeen)

console.log('\n===== P1: XSS 消毒 =====')
const xssPayload = '<img src=x onerror=alert(1)><script>alert(2)</script> onclick="x" javascript:alert(3)'
const px = (await req('POST', 'posts', { token, body: { title: xssPayload, message: xssPayload } })).data
const back = (await req('GET', `posts/${px._id}`, { token })).data?.post || {}
const msg = String(back.message || ''), title = String(back.title || '')
check('XSS payload sanitized in message', !msg.includes('<img') && !msg.includes('<script') && !msg.includes('onclick'), `stored=${msg.slice(0, 80)}`)
check('XSS payload sanitized in title', !title.includes('<img') && !title.includes('<script'), `stored=${title.slice(0, 80)}`)
await req('DELETE', `posts/${px._id}`, { token })

console.log('\n===== P2: 统一注册错误 + 帖子公开接口 =====')
const dup = await req('POST', 'user/signup', { body: { email, password: '123456', firstName: 'x' } })
const dupMsg = dup.data?.message || ''
check('duplicate email -> generic error (no email echo / no "already exists")',
  dup.status === 400 && !dupMsg.includes(email) && !dupMsg.includes('already exists'), `status=${dup.status}, msg="${dupMsg}"`)
const postsAnon = await req('GET', `posts?id=${uid}&page=1`)
check('feed requires auth now', postsAnon.status === 401, `status=${postsAnon.status}`)
const searchAnon = await req('GET', 'posts/search?searchQuery=test')
check('search requires auth now', searchAnon.status === 401, `status=${searchAnon.status}`)
const feedOk = await req('GET', `posts?id=${uid}&page=1`, { token })
check('feed works with auth', feedOk.status === 200, `status=${feedOk.status}`)

console.log('\n===== 回归：核心流程仍正常 =====')
// 用全新用户，避免被上面发帖限流测试影响
const su3 = (await req('POST', 'user/signup', { body: { email: uniq('reg'), password: 'reg123456', firstName: 'Reg' } })).data
const t3 = su3.token, u3 = su3.result._id
const p2 = (await req('POST', 'posts', { token: t3, body: { title: 'regression', message: 'still works' } })).data
const f2 = await req('GET', `posts/${p2._id}`, { token: t3 })
check('create + fetch post works', f2.status === 200 && f2.data?.post?.title === 'regression')
const like = await req('PATCH', `posts/${p2._id}/likePost`, { token: t3 })
check('like post works', like.status === 200)
await req('DELETE', `posts/${p2._id}`, { token: t3 })
const sug = await req('GET', 'user/getSug', { token: t3 })
check('getSug works', sug.status === 200)
const upd = await req('PATCH', 'user/update', { token: t3, body: { name: 'FixedUser', bio: 'updated' } })
check('update profile works (no password in response)', upd.status === 200 && !(upd.data?.data?.password))

console.log(`\n========== RESULT: ${closed} CLOSED, ${open} STILL OPEN ==========`)
if (notes.length) console.log(notes.join('\n'))
process.exit(open ? 1 : 0)
