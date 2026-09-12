package realtime

import (
	"log"
	"realTimeChat/servegrpc"
	"sync"
	"time"

	"github.com/gofiber/websocket/v2"
)

// writeTimeout bounds a single outbound write so one stalled client cannot hold
// up the broadcast for everybody else.
const writeTimeout = 5 * time.Second

type Message struct {
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Content  string `json:"content"`
}

// hubConn wraps one connection and serialises writes on it.
//
// The previous implementation wrote to connections while holding the global
// mutex; a single slow peer could therefore block all chat traffic.
type hubConn struct {
	c  *websocket.Conn
	mu sync.Mutex
}

func (hc *hubConn) writeJSON(v interface{}) error {
	// A nil connection is tolerated so the hub's bookkeeping (friend lists,
	// online/offline broadcasts) can be exercised in unit tests without a real
	// websocket upgrade.
	if hc == nil || hc.c == nil {
		return nil
	}
	hc.mu.Lock()
	defer hc.mu.Unlock()
	_ = hc.c.SetWriteDeadline(time.Now().Add(writeTimeout))
	return hc.c.WriteJSON(v)
}

// Hub owns the set of live connections and the "who is online" view.
//
// It replaces ConnectionManager, which had three structural defects:
//
//  1. isFriend was called *while holding the global lock* and blocked on a gRPC
//     round trip — once per online user, sequentially, with the lock held. Any
//     slow lookup stalled every chat delivery. The friend list is now fetched
//     once per connect, and no network call happens under the lock.
//  2. Its background goroutine mutated onlineFriends/connections *without* the
//     lock, racing with every other method. A concurrent map write is a fatal
//     runtime error in Go, not a recoverable one.
//  3. connections was map[string]*websocket.Conn, so a second login silently
//     kicked the first device off. Each user may now have several connections.
type Hub struct {
	mu             sync.Mutex
	conns          map[string]map[*hubConn]struct{}
	onlineFriends  map[string][]string
	getUserFriends func(string) <-chan []string
}

// NewHub creates an empty hub. getUserFriends may be nil (no friend lookups).
func NewHub(getUserFriends func(string) <-chan []string) *Hub {
	return &Hub{
		conns:          make(map[string]map[*hubConn]struct{}),
		onlineFriends:  make(map[string][]string),
		getUserFriends: getUserFriends,
	}
}

// friendIDs fetches the friend list. It performs a gRPC round trip and must
// never be called while holding h.mu.
func (h *Hub) friendIDs(userID string) []string {
	if h.getUserFriends == nil {
		return nil
	}
	ch := h.getUserFriends(userID)
	if ch == nil {
		return nil
	}
	var friends []string
	for batch := range ch {
		friends = append(friends, batch...)
	}
	return friends
}

// AddConnection registers a connection for userID and returns its handle.
//
// Side effects, matching the previous behaviour: the connecting user is told
// which of their friends are already online, and those friends are told the
// user just came online. Writes happen outside the lock.
func (h *Hub) AddConnection(userID string, c *websocket.Conn) *hubConn {
	hc := &hubConn{c: c}

	h.mu.Lock()
	if h.conns[userID] == nil {
		h.conns[userID] = make(map[*hubConn]struct{})
	}
	h.conns[userID][hc] = struct{}{}
	if _, ok := h.onlineFriends[userID]; !ok {
		h.onlineFriends[userID] = []string{}
	}
	others := make([]string, 0, len(h.conns))
	for id := range h.conns {
		if id != userID {
			others = append(others, id)
		}
	}
	h.mu.Unlock()

	// One lookup answers every friendship question (the old code asked once per
	// online user, each time holding the lock).
	friends := h.friendIDs(userID)
	friendSet := make(map[string]bool, len(friends))
	for _, f := range friends {
		friendSet[f] = true
	}

	type update struct {
		userID string
		list   []string
	}
	var updates []update

	h.mu.Lock()
	// Which of the new user's friends are online right now.
	online := make([]string, 0, len(friends))
	for _, f := range friends {
		if len(h.conns[f]) > 0 {
			online = append(online, f)
		}
	}
	h.onlineFriends[userID] = online
	updates = append(updates, update{userID: userID, list: append([]string(nil), online...)})

	// Tell each online friend that userID is now online.
	for _, other := range others {
		if !friendSet[other] {
			continue
		}
		h.onlineFriends[other] = appendUnique(h.onlineFriends[other], userID)
		updates = append(updates, update{userID: other, list: append([]string(nil), h.onlineFriends[other]...)})
	}
	h.mu.Unlock()

	for _, u := range updates {
		h.deliver(u.userID, u.list)
	}
	return hc
}

// RemoveConnection deregisters one connection. The offline broadcast only
// happens once the user's last connection is gone, so a second device closing
// does not make the user look offline.
func (h *Hub) RemoveConnection(userID string, hc *hubConn) {
	type update struct {
		userID string
		list   []string
	}
	var updates []update

	h.mu.Lock()
	if subs, ok := h.conns[userID]; ok {
		delete(subs, hc)
		if len(subs) == 0 {
			delete(h.conns, userID)
		}
	}
	if len(h.conns[userID]) == 0 {
		delete(h.onlineFriends, userID)
		for friendID, list := range h.onlineFriends {
			if idx := indexOf(list, userID); idx >= 0 {
				h.onlineFriends[friendID] = append(list[:idx], list[idx+1:]...)
				updates = append(updates, update{friendID, append([]string(nil), h.onlineFriends[friendID]...)})
			}
		}
	}
	h.mu.Unlock()

	for _, u := range updates {
		h.deliver(u.userID, u.list)
	}
}

// SendToReceiver persists the message and then pushes it to any live connection
// of the receiver.
//
// Note on ordering: the message is persisted first so a client that marks the
// thread as read cannot race ahead of the unread record. Also note that
// persistence now happens even when the receiver is offline — the old code
// dropped those messages on the floor (it only wrote to the database inside the
// "receiver is connected" branch), which silently lost data.
func (h *Hub) SendToReceiver(msg Message) {
	h.mu.Lock()
	subs := make([]*hubConn, 0, len(h.conns[msg.Receiver]))
	for hc := range h.conns[msg.Receiver] {
		subs = append(subs, hc)
	}
	h.mu.Unlock()

	if err := servegrpc.SendMessageClient(msg.Sender, msg.Receiver, msg.Content); err != nil {
		log.Printf("Error saving message to gRPC for %s: %v", msg.Receiver, err)
		return
	}

	if len(subs) == 0 {
		// Persisted for later delivery via the REST API; nothing to push.
		log.Printf("Receiver %s not online, message stored", msg.Receiver)
		return
	}
	for _, hc := range subs {
		if err := hc.writeJSON(msg); err != nil {
			log.Printf("Error sending message to %s : %v", msg.Receiver, err)
		}
	}
}

// deliver pushes the online-friend list to every connection of userID.
func (h *Hub) deliver(userID string, list []string) {
	h.mu.Lock()
	subs := make([]*hubConn, 0, len(h.conns[userID]))
	for hc := range h.conns[userID] {
		subs = append(subs, hc)
	}
	h.mu.Unlock()

	for _, hc := range subs {
		if err := hc.writeJSON(map[string]interface{}{"onlineFriends": list}); err != nil {
			log.Printf("Error notifying %s online friends: %v", userID, err)
		}
	}
}

// OnlineFriends returns a copy of the cached online-friend list (tests).
func (h *Hub) OnlineFriends(userID string) []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]string(nil), h.onlineFriends[userID]...)
}

// Count reports how many connections a user currently has (tests/logging).
func (h *Hub) Count(userID string) int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.conns[userID])
}

func appendUnique(list []string, v string) []string {
	if indexOf(list, v) >= 0 {
		return list
	}
	return append(list, v)
}

func indexOf(list []string, v string) int {
	for i, item := range list {
		if item == v {
			return i
		}
	}
	return -1
}
