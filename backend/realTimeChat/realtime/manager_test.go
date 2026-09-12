package realtime

import (
	"sync"
	"testing"
)

// friends 返回一个静态好友表，替代真实 gRPC 查询（真实实现每次都走一次 RPC）。
func friends(table map[string][]string) func(string) <-chan []string {
	return func(userID string) <-chan []string {
		ch := make(chan []string, 1)
		ch <- table[userID]
		close(ch)
		return ch
	}
}

// TestHubTracksMultipleConnectionsPerUser 覆盖旧实现的缺陷 3：
// 原来是 map[string]*websocket.Conn，同一个人第二端登录会把第一端顶掉。
func TestHubTracksMultipleConnectionsPerUser(t *testing.T) {
	h := NewHub(friends(nil))

	a1 := h.AddConnection("A", nil)
	a2 := h.AddConnection("A", nil)

	if got := h.Count("A"); got != 2 {
		t.Fatalf("A 的在线连接数 = %d，期望 2（多端应同时在线）", got)
	}

	// 关掉其中一端，另一端必须仍然在线
	h.RemoveConnection("A", a1)
	if got := h.Count("A"); got != 1 {
		t.Fatalf("关掉一端后 A 的连接数 = %d，期望 1", got)
	}
	h.RemoveConnection("A", a2)
	if got := h.Count("A"); got != 0 {
		t.Fatalf("全部关闭后 A 的连接数 = %d，期望 0", got)
	}
}

// TestHubExchangesOnlineFriendLists 覆盖上下线广播：
// A、B 互为好友，A 上线后双方都应看到彼此在线；A 下线后 B 的列表要清空。
func TestHubExchangesOnlineFriendLists(t *testing.T) {
	h := NewHub(friends(map[string][]string{
		"A": {"B"},
		"B": {"A"},
	}))

	hb := h.AddConnection("B", nil)
	ha := h.AddConnection("A", nil)

	if got := h.OnlineFriends("A"); len(got) != 1 || got[0] != "B" {
		t.Errorf("A 的在线好友 = %v，期望 [B]", got)
	}
	if got := h.OnlineFriends("B"); len(got) != 1 || got[0] != "A" {
		t.Errorf("B 的在线好友 = %v，期望 [A]", got)
	}

	h.RemoveConnection("A", ha)
	if got := h.OnlineFriends("B"); len(got) != 0 {
		t.Errorf("A 下线后 B 的在线好友 = %v，期望空", got)
	}

	// B 仍在线的场景下，重复移除 A 不应影响 B
	h.RemoveConnection("A", ha)
	if got := h.Count("B"); got != 1 {
		t.Errorf("B 的连接数 = %d，期望 1", got)
	}
	_ = hb
}

// TestHubIgnoresNonFriends 只有好友关系才应收到上下线通知。
func TestHubIgnoresNonFriends(t *testing.T) {
	h := NewHub(friends(map[string][]string{
		"A": {"C"}, // A 的好友只有 C
		"B": {},    // B 与 A 不是好友
	}))

	h.AddConnection("B", nil)
	h.AddConnection("A", nil)

	if got := h.OnlineFriends("B"); len(got) != 0 {
		t.Errorf("非好友之间的在线列表应为空，得到 %v", got)
	}
	if got := h.OnlineFriends("A"); len(got) != 0 {
		// A 的好友 C 并不在线
		t.Errorf("A 的在线好友应为空（C 未上线），得到 %v", got)
	}
}

// TestHubConcurrentAddRemove 是替换旧 ConnectionManager 的关键理由：
// 旧实现的后台 goroutine 不加锁地读写 onlineFriends/connections，
// 与其它方法构成数据竞争（Go 对并发 map 写是 fatal error，不可 recover）。
// 本用例在 -race 下运行；旧实现在这里会直接报 race 或 panic。
func TestHubConcurrentAddRemove(t *testing.T) {
	table := map[string][]string{
		"A": {"B", "C", "D"},
		"B": {"A"},
		"C": {"A"},
		"D": {"A"},
	}
	h := NewHub(friends(table))

	var wg sync.WaitGroup
	users := []string{"A", "B", "C", "D"}
	for round := 0; round < 20; round++ {
		for _, u := range users {
			wg.Add(1)
			go func(u string) {
				defer wg.Done()
				hc := h.AddConnection(u, nil)
				_ = h.OnlineFriends(u)
				_ = h.Count(u)
				h.RemoveConnection(u, hc)
			}(u)
		}
	}
	wg.Wait()

	for _, u := range users {
		if got := h.Count(u); got != 0 {
			t.Errorf("%s 的连接数 = %d，期望 0（全部已注销）", u, got)
		}
	}
}

func TestAppendUnique(t *testing.T) {
	list := []string{"a"}
	list = appendUnique(list, "a")
	if len(list) != 1 {
		t.Errorf("重复元素不应被追加: %v", list)
	}
	list = appendUnique(list, "b")
	if len(list) != 2 || list[1] != "b" {
		t.Errorf("新元素应被追加: %v", list)
	}
	if indexOf(list, "zzz") != -1 {
		t.Error("indexOf 对不存在的元素应返回 -1")
	}
}
