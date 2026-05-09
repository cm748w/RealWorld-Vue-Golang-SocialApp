package realtime

import (
	"log"
	"realTimeChat/servegrpc"
	"sync"

	"github.com/gofiber/websocket/v2"
)

type Message struct {
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Content  string `json:"content"`
}

type ConnectionManager struct {
	connections    map[string]*websocket.Conn
	onlineFriends  map[string][]string
	getUserFriends func(string) <-chan []string
	lock           sync.Mutex
}

func NewConnectionManager(getUserFriends func(string) <-chan []string) *ConnectionManager {
	if getUserFriends != nil {
		return &ConnectionManager{
			connections:    make(map[string]*websocket.Conn),
			onlineFriends:  make(map[string][]string),
			getUserFriends: getUserFriends,
		}
	}
	return nil
}

func (cm *ConnectionManager) AddConnection(userID string, conn *websocket.Conn) { // 用户上线
	cm.lock.Lock()
	defer cm.lock.Unlock()
	cm.connections[userID] = conn
	cm.onlineFriends[userID] = []string{}

	// Notify existing users about the new online friend
	for friendID := range cm.onlineFriends {
		if friendID != userID && cm.isFriend(userID, friendID) {
			cm.onlineFriends[friendID] = append(cm.onlineFriends[friendID], userID)
			err := cm.connections[friendID].WriteJSON(map[string]interface{}{
				"onlineFriends": cm.onlineFriends[friendID],
			})
			if err != nil {
				log.Printf("Error notifying %s about %s : %v", friendID, userID, err)
				return
			}
		}
	}

	// updata the online friends list for the new user
	go func() {
		for friends := range cm.getUserFriends(userID) {
			if friends == nil {
				continue
			}

			for _, friendID := range friends {
				if cm.connections[friendID] != nil {
					cm.onlineFriends[userID] = append(cm.onlineFriends[userID], friendID)
					err := cm.connections[userID].WriteJSON(map[string]interface{}{
						"onlineFriends": cm.onlineFriends[userID],
					})
					if err != nil {
						log.Printf("Error notifying %s about %s : %v", userID, friendID, err)
						return
					}
				}
			}
		}
	}()
}

func (cm *ConnectionManager) RemoveConnection(userID string) { // 用户下线
	cm.lock.Lock()
	defer cm.lock.Unlock()
	delete(cm.connections, userID)
	delete(cm.onlineFriends, userID)
	for friendID := range cm.onlineFriends {
		for i, id := range cm.onlineFriends[friendID] {
			if id == userID {
				cm.onlineFriends[friendID] = append(cm.onlineFriends[friendID][:i], cm.onlineFriends[friendID][i+1:]...)
				err := cm.connections[friendID].WriteJSON(map[string]interface{}{
					"onlineFriends": cm.onlineFriends[friendID],
				})
				if err != nil {
					log.Printf("Error notifying %s about %s : %v", friendID, userID, err)
				}
				break
			}
		}
	}
}

func (cm *ConnectionManager) SendToReceiver(msg Message) { // 发消息
	cm.lock.Lock()
	defer cm.lock.Unlock()
	if conn, ok := cm.connections[msg.Receiver]; ok {
		// Persist first, then deliver: avoids race where receiver marks as read before unread record exists.
		err := servegrpc.SendMessageClient(msg.Sender, msg.Receiver, msg.Content)
		if err != nil {
			log.Printf("Error saving message to gRPC for %s: %v", msg.Receiver, err)
			return
		}

		err = conn.WriteJSON(msg)
		if err != nil {
			log.Printf("Error sending message to %s : %v", msg.Receiver, err)
		}
	} else {
		log.Printf("Receiver %s not found ", msg.Receiver)
	}
}

// helper func is Friend
func (cm *ConnectionManager) isFriend(userID, friendID string) bool {
	friends := <-cm.getUserFriends(userID)
	if friends == nil {
		return false
	}
	for _, f := range friends {
		if f == friendID {
			return true
		}
	}
	return false
}
