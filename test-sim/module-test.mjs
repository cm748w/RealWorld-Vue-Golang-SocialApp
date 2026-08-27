// ============================================================================
//  Loads the REAL frontend api/index.js and utils/timing.js (as temp copies,
//  with baseURL pointed at the dev-server proxy) and verifies:
//   - auth request interceptor attaches the Bearer token
//   - in-flight dedup (concurrent identical GETs share one response)
//   - TTL cache hit / expiry
//   - token is part of the cache key (no cross-user leak)
//   - mutations clear the GET cache
//   - debounce / throttle timings behave as documented
//   - 429 retry under a concurrent burst
// ============================================================================
import { readFileSync, writeFileSync, unlinkSync, mkdirSync, existsSync, readdirSync, rmSync } from 'fs'
import { fileURLToPath, pathToFileURL } from 'url'
import path from 'path'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const repoRoot = path.resolve(__dirname, '..')
const frontendSrc = path.join(repoRoot, 'frontend', 'src')
const tmpDir = path.join(repoRoot, 'frontend', 'test-sim')
const apiCopyPath = path.join(tmpDir, '__api_copy.mjs')
const timingCopyPath = path.join(tmpDir, '__timing_copy.mjs')

// ---- minimal localStorage polyfill (browser API used by api/index.js) ----
const store = new Map()
globalThis.localStorage = {
  getItem: (k) => (store.has(k) ? store.get(k) : null),
  setItem: (k, v) => store.set(k, String(v)),
  removeItem: (k) => store.delete(k),
  clear: () => store.clear(),
}

let passed = 0
let failed = 0
const failures = []
function check(name, cond, extra = '') {
  if (cond) { passed++; console.log(`  ✅ ${name}${extra ? '  (' + extra + ')' : ''}`) }
  else { failed++; failures.push(name); console.log(`  ❌ ${name}${extra ? '  (' + extra + ')' : ''}`) }
}
const sleep = (ms) => new Promise((r) => setTimeout(r, ms))
const uniq = (p) => `${p}_${Date.now()}_${Math.floor(Math.random() * 1e6)}`

// ---- prepare temp copies ----
mkdirSync(tmpDir, { recursive: true })
let apiSource = readFileSync(path.join(frontendSrc, 'api', 'index.js'), 'utf8')
apiSource = apiSource.replace("baseURL: '/api/'", "baseURL: 'http://localhost:8080/api/'")
writeFileSync(apiCopyPath, apiSource)
writeFileSync(timingCopyPath, readFileSync(path.join(frontendSrc, 'utils', 'timing.js'), 'utf8'))

let exitCode = 0
try {
  const api = (await import(pathToFileURL(apiCopyPath).href)).default || await import(pathToFileURL(apiCopyPath).href)
  const timing = await import(pathToFileURL(timingCopyPath).href)
  const { debounce, throttle } = timing

  // ---- create two real users to test with ----
  const signup = async (email) => {
    const r = await fetch('http://localhost:8080/api/user/signup', {
      method: 'POST', headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ email, password: '123456', firstName: 'Mod', lastName: 'Tester' }),
    })
    return r.json()
  }
  const A = await signup(uniq('ma'))
  const B = await signup(uniq('mb'))
  const tokenA = A.token, idA = A.result._id
  const tokenB = B.token, idB = B.result._id
  const setProfile = (token, id) => localStorage.setItem('profile', JSON.stringify({ token, result: { _id: id } }))

  console.log('\n===== M1. AUTH INTERCEPTOR (through the real module) =====')
  setProfile(tokenA, idA)
  let sug = await api.getSugUser()
  check('M1.1 getSug succeeds with token (no 401)', sug.status === 200, `status=${sug.status}`)
  let self = await api.fetchUserProfile(idA)
  check('M1.2 fetchUserProfile(self) 200', self.status === 200, `status=${self.status}`)

  console.log('\n===== M2. IN-FLIGHT DEDUP =====')
  const five = await Promise.all([...Array(5)].map(() => api.fetchUserProfile(idB)))
  const allSame = five.every((r) => r === five[0])
  check('M2.1 5 concurrent identical GETs share one response object', allSame)
  check('M2.2 shared response is valid (200)', five[0].status === 200, `status=${five[0].status}`)

  console.log('\n===== M3. TTL CACHE =====')
  const s1 = await api.getSugUser()
  const s2 = await api.getSugUser()
  check('M3.1 repeat GET within TTL returns cached object', s1 === s2)
  const u1 = await api.fetchUnreadMessageSummary()
  const u2 = await api.fetchUnreadMessageSummary()
  check('M3.2 unread summary cached (3s TTL)', u1 === u2)
  await sleep(3200)
  const u3 = await api.fetchUnreadMessageSummary()
  check('M3.3 unread summary refetched after TTL expiry', u3 !== u2)

  console.log('\n===== M4. TOKEN IS PART OF CACHE KEY =====')
  const gA = await api.getSugUser()          // cached under tokenA key
  setProfile(tokenB, idB)
  const gB = await api.getSugUser()          // different token -> different key
  check('M4.1 different user does not reuse cache', gA !== gB && gB.status === 200)
  setProfile(tokenA, idA)

  console.log('\n===== M5. SELECTIVE CACHE INVALIDATION =====')
  // 造一条帖子用于 posts 族写操作
  const created = await api.createPost({ title: 'SelectiveTest', message: 'x' })
  const postId = created.data?._id || created._id
  check('M5.1 createPost ok (posts mutation)', !!postId)

  // 作者资料（user 族）先缓存
  const uA1 = await api.fetchUserProfile(idB)
  const uA2 = await api.fetchUserProfile(idB)
  check('M5.2 user profile cached', uA1 === uA2)

  // 点赞（posts 族写操作）—— 不应清掉 user 族缓存
  const likeRes = await api.likePost(postId)
  check('M5.3 likePost ok (posts mutation)', likeRes.status === 200, `status=${likeRes.status}`)
  const uA3 = await api.fetchUserProfile(idB)
  check('M5.4 user profile cache SURVIVES posts mutation', uA3 === uA1)

  // getSug（user 族）也缓存
  const g1 = await api.getSugUser()
  const g2 = await api.getSugUser()
  check('M5.5 getSug cached', g1 === g2)

  // 关注（user 族写操作）—— 应清掉 user 族缓存
  const fol = await api.following(idB)
  check('M5.6 following ok (user mutation)', fol.status === 200, `status=${fol.status}`)
  const uA4 = await api.fetchUserProfile(idB)
  const g3 = await api.getSugUser()
  check('M5.7 user profile refetched after user mutation', uA4 !== uA1)
  check('M5.8 getSug refetched after user mutation', g3 !== g1)

  if (postId) { await api.deletePost(postId) }

  console.log('\n===== M6. DEBOUNCE / THROTTLE TIMINGS =====')
  let dCalls = 0
  const d = debounce(() => { dCalls++ }, 120)
  d(); d(); d()
  await sleep(60); d()
  await sleep(200)
  check('M6.1 debounce fires exactly once after quiet window', dCalls === 1, `calls=${dCalls}`)

  let tCalls = 0
  const t = throttle(() => { tCalls++ }, 120, { trailing: false })
  t(); t(); t()
  check('M6.2 throttle leading fires immediately once', tCalls === 1, `calls=${tCalls}`)
  await sleep(150); t()
  check('M6.3 throttle allows again after window', tCalls === 2, `calls=${tCalls}`)

  console.log('\n===== M7. CONCURRENT BURST (429 observation) =====')
  const pages = await Promise.all(
    [...Array(12)].map((_, i) =>
      api.fetchPosts(idA, { page: i + 1 })
        .then((r) => ({ i: i + 1, ok: true, status: r.status }))
        .catch((e) => ({ i: i + 1, ok: false, status: e.response?.status || 'ERR' }))
    )
  )
  const statuses = pages.map((p) => `p${p.i}:${p.status}`).join(' ')
  console.log('     burst statuses:', statuses)
  const allOk = pages.every((p) => p.ok)
  check('M7.1 all burst requests succeed (429 auto-retried if any)', allOk, allOk ? '' : 'see statuses above')

  console.log(`\n========== MODULE RESULT: ${passed} passed, ${failed} failed ==========`)
  if (failures.length) console.log('Failed:', failures.join(' | '))
  exitCode = failed ? 1 : 0
} finally {
  // cleanup temp files (retry: Windows may hold the module file handle briefly)
  for (const p of [apiCopyPath, timingCopyPath]) {
    for (let i = 0; i < 5; i++) {
      try {
        if (existsSync(p)) unlinkSync(p)
        break
      } catch {
        await sleep(80)
      }
    }
  }
  try { rmSync(tmpDir, { recursive: true, force: true }) } catch { /* ignore */ }
}
process.exit(exitCode)
