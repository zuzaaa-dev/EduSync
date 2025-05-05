package ws

import (
	"EduSync/internal/repository"
	"EduSync/internal/util"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Client struct {
	conn   *websocket.Conn
	send   chan interface{}
	userID int
	hub    *Hub
}

// HandleWebSocket — Gin-handler для апгрейда HTTP → WebSocket
func HandleWebSocket(hub *Hub, mgr *util.JWTManager, tokenRepo repository.TokenRepository, log *logrus.Logger) gin.HandlerFunc {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	return func(c *gin.Context) {
		// 1) JWT из заголовка
		auth := c.GetHeader("Authorization")
		parts := strings.Split(auth, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "требуется авторизация"})
			return
		}
		claims, err := mgr.ParseJWT(parts[1], log)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "некорректный токен"})
			return
		}
		valid, _ := tokenRepo.IsValid(c.Request.Context(), parts[1])
		if !valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "токен отозван"})
			return
		}
		// 2) апгрейдим
		wsConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		client := &Client{
			conn:   wsConn,
			send:   make(chan interface{}, 256),
			userID: claims.ID,
			hub:    hub,
		}
		// 3) старт чита/пиши
		go client.writePump()
		client.readPump()
	}
}

// readPump читает команды от клиента (в т.ч. подписку)
func (c *Client) readPump() {
	defer c.conn.Close()
	for {
		var msg struct {
			Action string `json:"action"`
			ChatID int    `json:"chat_id"`
		}
		if err := c.conn.ReadJSON(&msg); err != nil {
			break
		}
		switch msg.Action {
		case "subscribe":
			room := fmt.Sprintf("chat_%d", msg.ChatID)
			c.hub.Subscribe(room, c)
		case "unsubscribe":
			room := fmt.Sprintf("chat_%d", msg.ChatID)
			c.hub.Unsubscribe(room, c)
		default:
			// игнорируем
		}
	}
}

// writePump шлёт все события клиенту
func (c *Client) writePump() {
	defer c.conn.Close()
	for payload := range c.send {
		c.conn.WriteJSON(payload)
	}
}
