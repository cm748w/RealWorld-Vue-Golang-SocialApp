package realtime

import (
	"sync"
	"time"

	"github.com/gofiber/websocket/v2"
)

// writeTimeout 限制单次下发的写超时：慢客户端不能拖住其它人的通知。
const writeTimeout = 5 * time.Second

// hubConn 包装一条 WS 连接，用独立互斥锁串行化**该连接**上的写操作。
// 这样全局 map 锁只在增删/快照时短暂持有，不再覆盖网络写——旧实现是
// 「持全局锁做 conn.WriteJSON」，一个卡住的客户端会阻塞所有人的通知下发。
type hubConn struct {
	c  *websocket.Conn
	mu sync.Mutex
}

// writeJSON 串行化写并带写超时，返回写入错误供调用方记录。
func (hc *hubConn) writeJSON(v interface{}) error {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	_ = hc.c.SetWriteDeadline(time.Now().Add(writeTimeout))
	return hc.c.WriteJSON(v)
}

// Hub 维护 userId → 在线连接集合的映射。
//
// 旧实现是 map[string]*websocket.Conn，同一用户两端登录会互相顶掉；
// 这里改为集合，多端可同时在线。
type Hub struct {
	mu    sync.Mutex
	conns map[string]map[*hubConn]struct{}
}

// NewHub 创建空的连接注册表。
func NewHub() *Hub {
	return &Hub{conns: make(map[string]map[*hubConn]struct{})}
}

// Add 注册一条连接并返回其句柄（供后续 Remove 使用）。
func (h *Hub) Add(userID string, c *websocket.Conn) *hubConn {
	hc := &hubConn{c: c}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.conns[userID] == nil {
		h.conns[userID] = make(map[*hubConn]struct{})
	}
	h.conns[userID][hc] = struct{}{}
	return hc
}

// Remove 注销一条连接，幂等；该用户没有连接时回收其条目。
func (h *Hub) Remove(userID string, hc *hubConn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	subs, ok := h.conns[userID]
	if !ok {
		return
	}
	delete(subs, hc)
	if len(subs) == 0 {
		delete(h.conns, userID)
	}
}

// Send 向某用户当前在线的全部连接下发 payload，返回逐条写入错误。
// 快照在锁内完成，实际网络写在锁外进行。
func (h *Hub) Send(userID string, payload interface{}) []error {
	h.mu.Lock()
	subs := make([]*hubConn, 0, len(h.conns[userID]))
	for hc := range h.conns[userID] {
		subs = append(subs, hc)
	}
	h.mu.Unlock()

	var errs []error
	for _, hc := range subs {
		if err := hc.writeJSON(payload); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// Count 返回某用户当前的连接数（仅用于日志/诊断）。
func (h *Hub) Count(userID string) int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.conns[userID])
}
