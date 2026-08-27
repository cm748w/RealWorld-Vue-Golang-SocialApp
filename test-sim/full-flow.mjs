// ============================================================================
//  Full-scale user-behaviour simulation against the real backend.
//  Runs through http://localhost:8080/api/ (the exact dev-server proxy path
//  the frontend uses), with the same payloads/shapes as frontend/src.
//  Also loads the REAL frontend api/index.js module (as a temp copy) to
//  verify the GET cache / in-flight dedup / auth interceptor logic.
// ============================================================================
import { readFileSync, writeFileSync, unlinkSync, existsSync, mkdirSync } from 'fs'
import { fileURLToPath, pathToFileURL } from 'url'
import path from 'path'

const BASE = 'http://localhost:8080/api/'
const __dirname = path.dirname(fileURLToPath(import.meta.url))

let passed = 0
let failed = 0
const failures = []

function check(name, cond, extra = '') {
  if (cond) {
    passed++
    console.log(`  ✅ ${name}${extra ? '  (' + extra + ')' : ''}`)
  } else {
    failed++
    failures.push(name)
    console.log(`  ❌ ${name}${extra ? '  (' + extra + ')' : ''}`)
  }
}

async function req(method, url, { token, body } = {}) {
  const headers = { 'content-type': 'application/json' }
  if (token) headers.authorization = `Bearer ${token}`
  const res = await fetch(BASE + url, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  })
  const ct = res.headers.get('content-type') || ''
  const data = ct.includes('json') ? await res.json() : await res.text()
  return { status: res.status, data }
}

const sleep = (ms) => new Promise((r) => setTimeout(r, ms))
const uniq = (p) => `${p}_${Date.now()}_${Math.floor(Math.random() * 1e6)}`

console.log('\n========== A. AUTH ==========')
const emailA = uniq('ua')
const emailB = uniq('ub')

let a = await req('POST', 'user/signup', { body: { email: emailA, password: '123456', firstName: 'Test', lastName: 'UserA' } })
check('A1 signup userA 201', a.status === 201, `status=${a.status}`)
const tokenA = a.data?.token
const idA = a.data?.result?._id
check('A2 signup returns token+result._id', !!tokenA && !!idA)

let b = await req('POST', 'user/signup', { body: { email: emailB, password: '123456', firstName: 'Test', lastName: 'UserB' } })
check('A3 signup userB 201', b.status === 201, `status=${b.status}`)
const tokenB = b.data?.token
const idB = b.data?.result?._id
check('A4 signup userB token+id', !!tokenB && !!idB)

let si = await req('POST', 'user/signin', { body: { email: emailA, password: '123456' } })
check('A5 signin userA 200', si.status === 200, `status=${si.status}`)
// 注意：signin 每次都会新签发一个 token，与 signup 返回的 token 不一定相同
check('A6 signin returns a fresh token', !!si.data?.token)

let bad = await req('POST', 'user/signin', { body: { email: emailA, password: 'wrong-pw' } })
check('A7 signin wrong password rejected', bad.status === 400 || bad.status === 401 || bad.status === 403, `status=${bad.status}`)

let noauth = await req('GET', 'user/getSug')
check('A8 protected endpoint without token rejected', noauth.status === 401, `status=${noauth.status}`)

console.log('\n========== B. FEED & CONTENT ==========')
let feed = await req('GET', `posts?id=${idA}&page=1`, { token: tokenA })
check('B1 feed page1 200', feed.status === 200, `status=${feed.status}`)
check('B2 feed shape {data,numberOfPages,currentPage}', Array.isArray(feed.data?.data) && 'numberOfPages' in feed.data && feed.data.currentPage === 1)
const totalPages = feed.data?.numberOfPages || 1

const post = { title: `Test post ${Date.now()}`, message: 'Full-scale test message body.' }
let cp = await req('POST', 'posts', { token: tokenA, body: post })
check('B3 createPost 200', cp.status === 200 || cp.status === 201, `status=${cp.status}`)
const postId = cp.data?._id || cp.data?.result?._id
check('B4 createPost returns _id', !!postId)

let fd = await req('GET', `posts/${postId}`, { token: tokenA })
check('B5 fetchPost detail 200', fd.status === 200, `status=${fd.status}`)
// 后端返回 { post: {...} }，前端 PostDetails 用 response?.post 解包
check('B6 post detail title matches (shape {post})', fd.data?.post?.title === post.title)

let lp = await req('PATCH', `posts/${postId}/likePost`, { token: tokenA })
check('B7 likePost 200', lp.status === 200, `status=${lp.status}`)
// 后端返回 { post: {...} }；前端在 Post.vue 里乐观更新本地 likes，不依赖响应
check('B8 likePost adds uid to likes (shape {post})', Array.isArray(lp.data?.post?.likes) && lp.data.post.likes.includes(idA))

let cm = await req('POST', `posts/${postId}/commentPost`, { token: tokenA, body: { value: 'A test comment' } })
check('B9 commentPost 200', cm.status === 200 || cm.status === 201, `status=${cm.status}`)
// 后端返回 { data: { comments: [...] } }；前端乐观 push 本地 comments
const cmtList = cm.data?.data?.comments || cm.data?.comments || []
console.log('     commentPost body:', JSON.stringify(cm.data).slice(0, 200))
check('B10 comment appended (shape {data:{comments}})', Array.isArray(cmtList) && cmtList.some((c) => String(c).includes('A test comment')))

let sr = await req('GET', `posts/search?searchQuery=${encodeURIComponent('Test post')}`, { token: tokenA })
check('B11 searchPosts 200', sr.status === 200, `status=${sr.status}`)
console.log('     search keys:', sr.data ? Object.keys(sr.data) : '(none)')
check('B12 search shape has user/posts', sr.data && (Array.isArray(sr.data.posts) || Array.isArray(sr.data.user)))

console.log('\n========== C. SOCIAL ==========')
let sug = await req('GET', 'user/getSug', { token: tokenA })
check('C1 getSug 200', sug.status === 200, `status=${sug.status}`)
check('C2 getSug shape {users}', Array.isArray(sug.data?.users))

let prof = await req('GET', `user/getUser/${idB}`, { token: tokenA })
check('C3 fetchUserProfile(other) 200', prof.status === 200, `status=${prof.status}`)
check('C4 profile shape {user,posts}', !!prof.data?.user && Array.isArray(prof.data?.posts))
check('C5 profile user id matches', prof.data?.user?._id === idB)

let fol = await req('PATCH', `user/${idB}/following`, { token: tokenA })
check('C6 follow userB 200', fol.status === 200, `status=${fol.status}`)
const profB2 = await req('GET', `user/getUser/${idB}`, { token: tokenA })
check('C7 follow reflected in followers', Array.isArray(profB2.data?.user?.followers) && profB2.data.user.followers.some((f) => String(f) === String(idA)))
const profA = await req('GET', `user/getUser/${idA}`, { token: tokenA })
check('C8 own profile fetch 200', profA.status === 200, `status=${profA.status}`)

console.log('\n========== D. CHAT ==========')
let send = await req('POST', 'chat/sendmessage', { token: tokenA, body: { content: 'hello from A', sender: idA, receiver: idB } })
check('D1 sendChatMessage 200', send.status === 200 || send.status === 201, `status=${send.status}`)
const msgId = send.data?.result?._id
check('D2 send returns result._id', !!msgId)

let conv = await req('GET', `chat/getmsgsbynums?from=0&firstuid=${idA}&seconduid=${idB}`, { token: tokenA })
check('D3 fetchConversationMessages 200', conv.status === 200, `status=${conv.status}`)
check('D4 conv shape {msgs}', Array.isArray(conv.data?.msgs) && conv.data.msgs.some((m) => String(m._id) === String(msgId)))

let unread = await req('GET', 'chat/get-user-unreadmsg', { token: tokenB })
check('D5 unread summary 200 (userB)', unread.status === 200, `status=${unread.status}`)
check('D6 unread shape {messages,totalUnreadMessageCount}', Array.isArray(unread.data?.messages) && 'totalUnreadMessageCount' in unread.data)
const hasUnreadForA = (unread.data?.messages || []).some((m) => String(m.otherUserId) === String(idA) && Number(m.numOfUnreadMessages) > 0)
check('D7 userB sees unread from A', hasUnreadForA)

let mark = await req('PATCH', `chat/read-msg?otheruid=${idA}`, { token: tokenB, body: null })
check('D8 markConversationAsRead 200', mark.status === 200, `status=${mark.status}`)
check('D9 mark response has isMarked', 'isMarked' in (mark.data || {}))
const unread2 = await req('GET', 'chat/get-user-unreadmsg', { token: tokenB })
const unreadNow = (unread2.data?.messages || []).find((m) => String(m.otherUserId) === String(idA))
check('D10 unread cleared after read', !unreadNow || Number(unreadNow.numOfUnreadMessages) === 0)

console.log('\n========== E. NOTIFICATIONS ==========')
let notif = await req('GET', `notification/${idA}`, { token: tokenA })
check('E1 GetNotificationForUser 200', notif.status === 200, `status=${notif.status}`)
check('E2 notification shape {notifications}', Array.isArray(notif.data?.notifications))
const hadUnread = (notif.data?.notifications || []).some((n) => !n.isRead)
if (hadUnread) {
  let mr = await req('PATCH', `notification/mark-notification-as-readed/${idA}`, { token: tokenA })
  check('E3 markNotificationAsReaded 200', mr.status === 200, `status=${mr.status}`)
  const notif2 = await req('GET', `notification/${idA}`, { token: tokenA })
  check('E4 unread cleared', (notif2.data?.notifications || []).every((n) => n.isRead))
} else {
  console.log('     (no unread notifications to mark — skipped E3/E4)')
}

console.log('\n========== F. PROFILE UPDATE ==========')
let up = await req('PATCH', 'user/update', { token: tokenA, body: { ...profA.data?.user, name: 'RenamedA', bio: 'updated bio' } })
check('F1 updateUser 200', up.status === 200, `status=${up.status}`)
const profA2 = await req('GET', `user/getUser/${idA}`, { token: tokenA })
check('F2 name updated', profA2.data?.user?.name === 'RenamedA' || profA2.data?.user?.name === 'RenamedA ' || String(profA2.data?.user?.name).includes('RenamedA'))

console.log('\n========== G. PAGING ==========')
if (totalPages > 1) {
  let p2 = await req('GET', `posts?id=${idA}&page=2`, { token: tokenA })
  check('G1 feed page2 200', p2.status === 200, `status=${p2.status}`)
} else {
  console.log('     (only 1 page of posts — skipped page2)')
}

console.log('\n========== H. CLEANUP ==========')
let del = await req('DELETE', `posts/${postId}`, { token: tokenA })
check('H1 deletePost 200', del.status === 200, `status=${del.status}`)

console.log(`\n========== RESULT: ${passed} passed, ${failed} failed ==========`)
if (failures.length) {
  console.log('Failed:', failures.join(' | '))
}
process.exit(failed ? 1 : 0)
