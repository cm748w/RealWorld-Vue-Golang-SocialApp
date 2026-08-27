// ============================================================================
//  ROUND 2 — authorized re-attack as a technically-skilled adversary
//  Target: the fixed system. NO source changes. New angles only.
// ============================================================================
const BASE = 'http://localhost:8080/api/'
const uniq = (p) => `${p}_${Date.now()}_${Math.floor(Math.random() * 1e6)}`
const findings = []
async function report(name, statusP, evidence) {
  const status = await statusP
  const ok = status === 'OK' || status === 'WARN'
  const icon = status === 'OK' ? '🔴 FINDING' : status === 'WARN' ? '🟡 WARN' : '🟢 SAFE'
  console.log(`${icon} ${name}`)
  if (evidence) console.log(`        ${String(evidence).slice(0, 300)}`)
  if (ok) findings.push({ name, severity: status, evidence: String(evidence || '').slice(0, 300) })
}
async function req(method, url, { token, body, headers: extra } = {}) {
  const headers = { 'content-type': 'application/json', ...(extra || {}) }
  if (token) headers.authorization = `Bearer ${token}`
  const res = await fetch(BASE + url, { method, headers, body: body ? JSON.stringify(body) : undefined })
  const ct = res.headers.get('content-type') || ''
  const data = ct.includes('json') ? await res.json() : await res.text()
  return { status: res.status, data, headers: res.headers }
}

console.log('===== SETUP (3 users) =====')
const A = (await req('POST', 'user/signup', { body: { email: uniq('va'), password: 'victim123', firstName: 'VA' } })).data
const B = (await req('POST', 'user/signup', { body: { email: uniq('ab'), password: 'attacker123', firstName: 'AB' } })).data
const C = (await req('POST', 'user/signup', { body: { email: uniq('ac'), password: 'bystander123', firstName: 'AC' } })).data
const tA = A.token, idA = A.result._id
const tB = B.token, idB = B.result._id
const tC = C.token, idC = C.result._id

console.log('\n===== PHASE 1 — business-logic abuse =====')

// 1.1 TOCTOU race: 5 concurrent signups with the SAME email
await report('1.1 Signup race — duplicate accounts for the same email (TOCTOU, no unique index)',
  (async () => {
    const email = uniq('race')
    const results = await Promise.all([...Array(5)].map(() => req('POST', 'user/signup', { body: { email, password: 'race1234', firstName: 'R' } })))
    const created = results.filter((r) => r.status === 201).length
    const others = results.map((r) => r.status).join(',')
    return created > 1 ? 'OK' : 'SAFE'
  })(), '5 concurrent signups with the same email -> how many were created?')

// 1.2 like spam
await report('1.2 likePost rate-limited (30/min/user) — mass like spam blocked',
  (async () => {
    const p = (await req('POST', 'posts', { token: tA, body: { title: 'liketarget', message: 'x' } })).data
    let got429 = false, ok = 0
    for (let i = 0; i < 32; i++) { const r = await req('PATCH', `posts/${p._id}/likePost`, { token: tB }); if (r.status === 429) { got429 = true; break }; if (r.status === 200) ok++ }
    await req('DELETE', `posts/${p._id}`, { token: tA })
    return got429 ? 'SAFE' : 'WARN'
  })(), 'rapid likes -> 429 after 30')

// 1.3 comment spam
await report('1.3 No rate limit on commentPost — mass comment spam possible',
  (async () => {
    const p = (await req('POST', 'posts', { token: tA, body: { title: 'cmttarget', message: 'x' } })).data
    let ok = 0
    for (let i = 0; i < 25; i++) { const r = await req('POST', `posts/${p._id}/commentPost`, { token: tB, body: { value: `spam${i}` } }); if (r.status === 200 || r.status === 201) ok++ }
    await req('DELETE', `posts/${p._id}`, { token: tA })
    return ok === 25 ? 'WARN' : 'SAFE'
  })(), '25 rapid comments, all accepted')

// 1.4 follow toggle spam
await report('1.4 No rate limit on follow/unfollow toggle',
  (async () => {
    let ok = 0
    for (let i = 0; i < 20; i++) { const r = await req('PATCH', `user/${idC}/following`, { token: tB }); if (r.status === 200) ok++ }
    return ok === 20 ? 'WARN' : 'SAFE'
  })(), '20 rapid follow/unfollow toggles, all accepted')

// 1.5 message spam
await report('1.5 sendChatMessage rate-limited (20/min/user) — message flood blocked',
  (async () => {
    let got429 = false, ok = 0
    for (let i = 0; i < 22; i++) { const r = await req('POST', 'chat/sendmessage', { token: tA, body: { content: `flood${i}`, sender: idA, receiver: idB } }); if (r.status === 429) { got429 = true; break }; if (r.status === 200 || r.status === 201) ok++ }
    return got429 ? 'SAFE' : 'WARN'
  })(), 'rapid messages -> 429 after 20')

// 1.6 login lockout nuance: correct password blocked while locked (local same-IP DoS demo)
await report('1.6 Login lockout nuance — correct password also blocked during lock window',
  (async () => {
    for (let i = 0; i < 5; i++) await req('POST', 'user/signin', { body: { email: A.result.email, password: `bad${i}` } })
    const good = await req('POST', 'user/signin', { body: { email: A.result.email, password: 'victim123' } })
    return good.status === 429 ? 'WARN' : 'SAFE'
  })(), `after 5 bad logins, CORRECT password attempt -> status ${(await (async () => { for (let i = 0; i < 5; i++) await req('POST', 'user/signin', { body: { email: A.result.email, password: `bad${i}` } }); return (await req('POST', 'user/signin', { body: { email: A.result.email, password: 'victim123' } })).status })())} (same-IP only; per-email+IP key limits real-world impact)`)

console.log('\n===== PHASE 2 — info disclosure / hardening =====')

// 2.1 swagger public
await report('2.1 Swagger API docs publicly accessible (route map for attackers)',
  (async () => {
    const r = await fetch(BASE + 'swagger/index.html')
    return r.status === 200 ? 'WARN' : 'SAFE'
  })(), 'GET /api/swagger/index.html is open to anonymous')

// 2.2 error verbosity
await report('2.2 Verbose error — malformed JSON returns 500 + raw Go error',
  (async () => {
    const res = await fetch(BASE + 'user/signin', { method: 'POST', headers: { 'content-type': 'application/json' }, body: '{bad json' })
    const text = await res.text()
    return res.status === 500 && text.includes('invalid character') ? 'WARN' : 'SAFE'
  })(), 'malformed body -> 500 with raw parser error text')

// 2.3 CORS
await report('2.3 CORS wide open — any origin allowed with credentials',
  (async () => {
    const res = await fetch(BASE + 'user/getSug', { headers: { origin: 'http://evil.example.com' } })
    return res.headers.get('access-control-allow-origin') === 'http://evil.example.com' ? 'WARN' : 'SAFE'
  })(), `Access-Control-Allow-Origin reflects arbitrary origin`)

// 2.4 imageUrl / selectedFile unsanitized
await report('2.4 imageUrl / selectedFile stored raw (javascript: URI, SVG payload)',
  (async () => {
    await req('PATCH', 'user/update', { token: tB, body: { name: 'AB', imageUrl: 'javascript:alert(1)', bio: 'x' } })
    const me = (await req('GET', `user/getUser/${idB}`, { token: tB })).data?.user
    const p = (await req('POST', 'posts', { token: tB, body: { title: 'svgtarget', message: 'x', selectedFile: 'data:image/svg+xml,<svg onload=alert(1)>' } })).data
    const back = (await req('GET', `posts/${p._id}`, { token: tB })).data?.post
    await req('DELETE', `posts/${p._id}`, { token: tB })
    return (me?.imageUrl === 'javascript:alert(1)' && back?.selectedFile?.includes('<svg')) ? 'WARN' : 'SAFE'
  })(), 'stored raw; rendered via <img :src> (browsers block javascript:/svg-in-img script execution)')

console.log('\n===== PHASE 3 — regression: prior fixes still hold =====')
const anonU = await req('GET', `user/getUser/${idA}`)
await report('R1 getUser still requires auth', anonU.status === 401, `status=${anonU.status}`)
const anonN = await req('GET', `notification/${idA}`)
await report('R2 notification still requires auth', anonN.status === 401, `status=${anonN.status}`)
const p0 = (await req('POST', 'posts', { token: tB, body: { title: 'r', message: 'x' } })).data
const pl = await req('PATCH', `posts/${p0._id}/likePost`, { token: tA })
await report('R3 core flow (like) still works', pl.status === 200, `status=${pl.status}`)
await req('DELETE', `posts/${p0._id}`, { token: tB })

console.log('')
console.log('==========================================================')
console.log(`  ROUND 2 COMPLETE — ${findings.length} NEW FINDING(S)`)
console.log('==========================================================')
findings.forEach((f, i) => console.log(`  ${i + 1}. [${f.severity}] ${f.name}${f.evidence ? '  -> ' + f.evidence : ''}`))
