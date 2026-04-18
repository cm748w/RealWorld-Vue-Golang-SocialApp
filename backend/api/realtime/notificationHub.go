package realtime

import "sync"

// NotificationHub 按用户维护 SSE 订阅连接。
type NotificationHub struct {
	mu      sync.RWMutex
	clients map[string]map[chan string]struct{}
}

func NewNotificationHub() *NotificationHub {
	return &NotificationHub{
		clients: make(map[string]map[chan string]struct{}),
	}
}

var Notifications = NewNotificationHub()

// Subscribe 返回指定用户的事件通道和取消订阅函数。
func (h *NotificationHub) Subscribe(userID string) (chan string, func()) {
	ch := make(chan string, 16)

	h.mu.Lock()
	if h.clients[userID] == nil {
		h.clients[userID] = make(map[chan string]struct{})
	}
	h.clients[userID][ch] = struct{}{}
	h.mu.Unlock()

	cancel := func() {
		h.mu.Lock()
		if subs, ok := h.clients[userID]; ok {
			if _, exists := subs[ch]; exists {
				delete(subs, ch)
				close(ch)
			}
			if len(subs) == 0 {
				delete(h.clients, userID)
			}
		}
		h.mu.Unlock()
	}

	return ch, cancel
}

// Publish 向指定用户的全部订阅连接广播消息。
func (h *NotificationHub) Publish(userID string, payload string) {
	h.mu.RLock()
	subs := h.clients[userID]
	for ch := range subs {
		select {
		case ch <- payload:
		default:
		}
	}
	h.mu.RUnlock()
}
