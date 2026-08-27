package controllers

import (
	"Server/middleware"
	"time"
)

// 全局限流器（进程内单例）
var (
	// 注册：每 IP 每分钟最多 10 次（防垃圾账号灌水 / 邮箱枚举）
	signupLimiter = middleware.NewRateLimiter(10, time.Minute)
	// 发帖：每用户每分钟最多 5 次
	postCreateLimiter = middleware.NewRateLimiter(5, time.Minute)
	// 点赞：每用户每分钟最多 30 次
	likeLimiter = middleware.NewRateLimiter(30, time.Minute)
	// 评论：每用户每分钟最多 10 次
	commentLimiter = middleware.NewRateLimiter(10, time.Minute)
	// 关注/取关：每用户每分钟最多 10 次
	followLimiter = middleware.NewRateLimiter(10, time.Minute)
	// 私信：每用户每分钟最多 20 条
	messageLimiter = middleware.NewRateLimiter(20, time.Minute)
	// 登录：每 IP 每分钟最多 20 次尝试
	loginIPLimiter = middleware.NewRateLimiter(20, time.Minute)
	// 登录失败锁定：每账号(+IP) 15 分钟内失败 5 次 → 锁定 1 分钟
	loginGuard = middleware.NewLoginGuard(5, time.Minute, 15*time.Minute)
)
