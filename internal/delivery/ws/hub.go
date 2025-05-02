package ws

import (
	"sync"
)

// Hub управляет всеми активными подключениями и румами.
type Hub struct {
	// комната → набор соединений
	rooms map[string]map[*Client]bool
	mu    sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		rooms: make(map[string]map[*Client]bool),
	}
}

// Subscribe добавляет клиента в комнату
func (h *Hub) Subscribe(room string, c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	conns := h.rooms[room]
	if conns == nil {
		conns = make(map[*Client]bool)
		h.rooms[room] = conns
	}
	conns[c] = true
}

// Unsubscribe удаляет клиента из комнаты
func (h *Hub) Unsubscribe(room string, c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conns := h.rooms[room]; conns != nil {
		delete(conns, c)
		if len(conns) == 0 {
			delete(h.rooms, room)
		}
	}
}

// Broadcast шлёт событие во все соединения комнаты
func (h *Hub) Broadcast(room, event string, data interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	payload := map[string]interface{}{
		"event": event,
		"data":  data,
	}
	for c := range h.rooms[room] {
		c.send <- payload
	}
}
