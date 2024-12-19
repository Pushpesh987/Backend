package notifications

import (
	"log"
	"sync"
	

	// "github.com/gofiber/fiber/v
	"github.com/gofiber/websocket/v2"
)

// Store WebSocket connections
var notificationClients = make(map[*websocket.Conn]bool)
var mu sync.Mutex
var notificationBroadcast = make(chan Notification)

// Notification structure
type Notification struct {
	UserID  string `json:"user_id"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

// Fiber WebSocket Handler
func NotificationWebSocketHandler(c *websocket.Conn) {
	mu.Lock()
	notificationClients[c] = true
	mu.Unlock()

	log.Println("New WebSocket client connected for notifications")

	defer func() {
		mu.Lock()
		delete(notificationClients, c)
		mu.Unlock()
		c.Close()
	}()

	// Keep connection open and listen for incoming messages (optional)
	for {
		_, _, err := c.ReadMessage()
		if err != nil {
			log.Println("WebSocket client disconnected:", err)
			break
		}
	}
}

// Broadcast Notifications
func BroadcastNotifications() {
	for {
		notification := <-notificationBroadcast
		mu.Lock()
		for client := range notificationClients {
			err := client.WriteJSON(notification)
			if err != nil {
				log.Println("Error sending notification:", err)
				client.Close()
				delete(notificationClients, client)
			}
		}
		mu.Unlock()
	}
}

// Function to trigger notifications
func SendNotification(userID, title, message string) {
	notification := Notification{
		UserID:  userID,
		Title:   title,
		Message: message,
	}
	notificationBroadcast <- notification
}

