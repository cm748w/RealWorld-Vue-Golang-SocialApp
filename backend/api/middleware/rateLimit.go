package middleware

import (
	"sync"
	"time"
)

// RateLimiter 一个简单的固定窗口限流器（按 key 计数）
type rateBucket struct {
	windowStart time.Time
	count       int
}

type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*rateBucket
	max     int
	window  time.Duration
}

// NewRateLimiter 创建限流器：每个 key 在 window 时间内最多 max 次
func NewRateLimiter(max int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		buckets: make(map[string]*rateBucket),
		max:     max,
		window:  window,
	}
}

// Allow 记录一次访问；窗口内超过上限返回 false
func (r *RateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	b, ok := r.buckets[key]
	if !ok || now.Sub(b.windowStart) >= r.window {
		r.buckets[key] = &rateBucket{windowStart: now, count: 1}
		return true
	}
	b.count++
	return b.count <= r.max
}

// ---------------------------------------------------------------------------

// failState 单账号/单来源的失败计数与锁定状态
type failState struct {
	first       time.Time
	count       int
	lockedUntil time.Time
}

// LoginGuard 登录失败锁定：窗口内失败超过阈值即锁定一段时间
type LoginGuard struct {
	mu      sync.Mutex
	fails   map[string]*failState
	maxFail int
	lockDur time.Duration
	window  time.Duration
}

// NewLoginGuard 创建登录守卫：窗口内最多 maxFail 次失败，超限锁定 lockDur
func NewLoginGuard(maxFail int, lockDur, window time.Duration) *LoginGuard {
	return &LoginGuard{
		fails:   make(map[string]*failState),
		maxFail: maxFail,
		lockDur: lockDur,
		window:  window,
	}
}

// CheckAndRecord 在登录失败时调用：已锁定返回 false；否则记录一次失败，超阈值则锁定
func (g *LoginGuard) CheckAndRecord(key string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now()
	st, ok := g.fails[key]
	if !ok {
		st = &failState{first: now}
		g.fails[key] = st
	}

	if now.Before(st.lockedUntil) {
		return false
	}

	if now.Sub(st.first) >= g.window {
		st.first = now
		st.count = 0
	}

	st.count++
	if st.count >= g.maxFail {
		st.lockedUntil = now.Add(g.lockDur)
		st.count = 0
	}
	return true
}

// Reset 登录成功时清除失败记录
func (g *LoginGuard) Reset(key string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.fails, key)
}
