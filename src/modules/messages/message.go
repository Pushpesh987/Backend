package messages

import (
	"Backend/src/core/config"
	"Backend/src/core/database"
	"Backend/src/core/helpers"
	"Backend/src/core/models"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var communityConnections = make(map[string][]*websocket.Conn)
var mu sync.Mutex

func WebSocketHandler(c *fiber.Ctx) error {
	log.Println("Starting WebSocketHandler")

	userID, err := ExtractUserIDFromJWT(c)
	if err != nil {
		log.Println("Error extracting user_id:", err)
		return c.Status(fiber.StatusUnauthorized).SendString("Invalid token or missing Authorization header")
	}

	c.Locals("user_id", userID)

	if websocket.IsWebSocketUpgrade(c) {
		return c.Next()
	}

	return fiber.ErrUpgradeRequired
}

func ExtractUserIDFromJWT(c *fiber.Ctx) (string, error) {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		log.Println("Authorization header missing")
		return "", fmt.Errorf("authorization header missing")
	}

	if len(authHeader) < len("Bearer ") {
		log.Println("Authorization header is not in the correct format")
		return "", fmt.Errorf("invalid Authorization format")
	}

	tokenString := authHeader[len("Bearer "):]

	userID, err := validateJWT(tokenString)
	if err != nil {
		log.Println("Invalid token:", err)
		return "", fmt.Errorf("invalid token: %v", err)
	}

	return userID, nil
}

// func WebSocketConnHandler(conn *websocket.Conn) {
//     userIDStr := conn.Query("user_id")
//     if userIDStr == "" {
//         log.Println("user_id missing in WebSocket connection")
//         return
//     }

//     log.Println("user_id from query:", userIDStr)

//     userID, err := uuid.Parse(userIDStr)
//     if err != nil {
//         log.Println("Error parsing userID:", err)
//         return
//     }

//     log.Println("User ID parsed successfully:", userID)

//     communityIDStr := conn.Params("id")
//     communityID, err := strconv.Atoi(communityIDStr)
//     if err != nil {
//         log.Println("Error converting communityID to int:", err)
//         return
//     }

//     log.Println("WebSocket connection established for community:", communityID)

//     mu.Lock()
//     communityConnections[communityIDStr] = append(communityConnections[communityIDStr], conn)
//     mu.Unlock()

//     for {
//         msgType, msg, err := conn.ReadMessage()
//         if err != nil {
//             log.Println("Error reading message:", err)
//             break
//         }

//         log.Printf("Message received: %s\n", string(msg))

//         message := &models.Message{
//             CommunityID: communityID,
//             UserID:      userID,
//             Message:     string(msg),
//             CreatedAt:   time.Now(),
//         }

//         err = SendMessage(message)
//         if err != nil {
//             log.Println("Error saving message to database:", err)
//         }

//         mu.Lock()
//         for _, otherConn := range communityConnections[communityIDStr] {
//             if otherConn == conn {
//                 continue
//             }
//             if err := otherConn.WriteMessage(msgType, msg); err != nil {
//                 log.Println("Error sending message:", err)
//             }
//         }
//         mu.Unlock()
//     }

//     mu.Lock()
//     for i, ws := range communityConnections[communityIDStr] {
//         if ws == conn {
//             communityConnections[communityIDStr] = append(communityConnections[communityIDStr][:i], communityConnections[communityIDStr][i+1:]...)
//             break
//         }
//     }
//     mu.Unlock()

//     log.Println("WebSocket connection closed for community:", communityID)
// }

// func WebSocketConnHandler(conn *websocket.Conn) {

// 	userIDStr := conn.Query("user_id")
// 	if userIDStr == "" {
// 		log.Println("user_id missing in WebSocket connection")
// 		return
// 	}

// 	userID, err := uuid.Parse(userIDStr)
// 	if err != nil {
// 		log.Printf("Error parsing userID: %v", err)
// 		return
// 	}
// 	log.Printf("User ID: %v", userID)

// 	communityIDStr := conn.Params("id")
// 	communityID, err := strconv.Atoi(communityIDStr)
// 	if err != nil {
// 		log.Printf("Error converting communityID to int: %v", err)
// 		return
// 	}
// 	log.Printf("Community ID: %v", communityID)

// 	mu.Lock()
// 	log.Printf("Adding connection to community %v", communityIDStr)
// 	communityConnections[communityIDStr] = append(communityConnections[communityIDStr], conn)
// 	mu.Unlock()

// 	defer func() {
// 		log.Printf("Closing WebSocket connection for community: %v", communityID)
// 		conn.Close()
// 	}()

// 	for {
// 		msgType, msg, err := conn.ReadMessage()
// 		if err != nil {
// 			log.Printf("Error reading message: %v", err)
// 			break
// 		}

// 		log.Printf("Received message from user %v in community %v: Type: %v, Message: %s", userID, communityID, msgType, string(msg))

// 		message := &models.Message{
// 			CommunityID: communityID,
// 			UserID:      userID,
// 			Message:     string(msg),
// 			CreatedAt:   time.Now(),
// 		}
// 		log.Printf("Prepared message for database: %+v", message)

// 		go func() {
// 			err := SendMessage(message)
// 			if err != nil {
// 				log.Printf("Error saving message to database: %v", err)
// 			} else {
// 				log.Printf("Message saved to database: %+v", message)
// 			}
// 		}()

// 		mu.Lock()
// 		log.Printf("Broadcasting message to other connections in community %v", communityIDStr)
// 		for _, otherConn := range communityConnections[communityIDStr] {
// 			if otherConn == conn {
// 				continue
// 			}
// 			if err := otherConn.WriteMessage(msgType, msg); err != nil {
// 				log.Printf("Error sending message to %v: %v", otherConn.RemoteAddr(), err)
// 			} else {
// 				log.Printf("Message sent to %v", otherConn.RemoteAddr())
// 			}
// 		}
// 		mu.Unlock()
// 	}

// 	mu.Lock()
// 	log.Printf("Removing connection from community %v", communityIDStr)
// 	for i, ws := range communityConnections[communityIDStr] {
// 		if ws == conn {
// 			communityConnections[communityIDStr] = append(communityConnections[communityIDStr][:i], communityConnections[communityIDStr][i+1:]...)
// 			break
// 		}
// 	}
// 	mu.Unlock()

// 	log.Printf("WebSocket connection closed for community: %v", communityID)
// }

func WebSocketConnHandler(conn *websocket.Conn) {
	// Extract userID from the query parameters
	userIDStr := conn.Query("user_id")
	if userIDStr == "" {
		log.Println("user_id missing in WebSocket connection")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("Error parsing userID: %v", err)
		return
	}
	log.Printf("User ID: %v", userID)

	// Extract communityID from the URL parameters
	communityIDStr := conn.Params("id")
	communityID, err := strconv.Atoi(communityIDStr)
	if err != nil {
		log.Printf("Error converting communityID to int: %v", err)
		return
	}
	log.Printf("Community ID: %v", communityID)

	// Lock the shared map and add this connection to the communityConnections
	mu.Lock()
	log.Printf("Adding connection to community %v", communityIDStr)
	communityConnections[communityIDStr] = append(communityConnections[communityIDStr], conn)
	mu.Unlock()

	defer func() {
		log.Printf("Closing WebSocket connection for community: %v", communityID)
		conn.Close()
	}()

	// Loop to read messages from the WebSocket connection
	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		// Log the received message and message type
		log.Printf("Received message from user %v in community %v: Type: %v, Message: %s", userID, communityID, msgType, string(msg))

		// Fetch the username from the database using the userID
		username, err := GetUsernameByID(userID)
		if err != nil {
			log.Printf("Error fetching username: %v", err)
			username = "Unknown" // Fallback if username can't be fetched
		}

		// Create a Message struct and save it
		message := &models.Message{
			CommunityID: communityID,
			UserID:      userID,
			Message:     string(msg),
			CreatedAt:   time.Now(),
		}
		log.Printf("Prepared message for database: %+v", message)

		// Save the message to the database (async)
		go func() {
			err := SendMessage(message)
			if err != nil {
				log.Printf("Error saving message to database: %v", err)
			} else {
				log.Printf("Message saved to database: %+v", message)
			}
		}()

		// Broadcast the message to other connected clients in the community in JSON format
		mu.Lock()
		log.Printf("Broadcasting message to other connections in community %v", communityIDStr)
		for _, otherConn := range communityConnections[communityIDStr] {
			if otherConn == conn {
				continue
			}

			// Create a JSON message to broadcast
			broadcastMessage := map[string]interface{}{
				"username":    username,
				"message":     string(msg),
				"createdAt":   message.CreatedAt,
				"communityID": communityID,
			}

			// Convert the message to JSON
			jsonMessage, err := json.Marshal(broadcastMessage)
			if err != nil {
				log.Printf("Error marshalling message to JSON: %v", err)
				continue
			}

			// Send the JSON message to the other clients
			if err := otherConn.WriteMessage(msgType, jsonMessage); err != nil {
				log.Printf("Error sending message to %v: %v", otherConn.RemoteAddr(), err)
			} else {
				log.Printf("Message sent to %v", otherConn.RemoteAddr())
			}
		}
		mu.Unlock()
	}

	// Remove the connection from communityConnections after the connection is closed
	mu.Lock()
	log.Printf("Removing connection from community %v", communityIDStr)
	for i, ws := range communityConnections[communityIDStr] {
		if ws == conn {
			communityConnections[communityIDStr] = append(communityConnections[communityIDStr][:i], communityConnections[communityIDStr][i+1:]...)
			break
		}
	}
	mu.Unlock()

	log.Printf("WebSocket connection closed for community: %v", communityID)
}

func GetUsernameByID(userID uuid.UUID) (string, error) {
	db := database.DB

	var user models.User
	// Query your database using the userID to get the user's details
	err := db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		return "", fmt.Errorf("error fetching username: %v", err)
	}
	return user.Username, nil
}

func validateJWT(tokenString string) (string, error) {
	log.Println("Validating JWT token")
	jwtSecret := config.Config("JWT_SECRET")
	if jwtSecret == "" {
		log.Println("JWT_SECRET is not set")
		return "", fmt.Errorf("JWT_SECRET is not set")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Println("Unexpected signing method:", token.Header["alg"])
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		log.Println("Error parsing token:", err)
		return "", err
	}

	if !token.Valid {
		log.Println("Invalid or expired token")
		return "", fmt.Errorf("invalid or expired token")
	}

	log.Println("Token is valid")

	// Extract user_id from the token claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Println("Error extracting claims from token")
		return "", fmt.Errorf("error extracting claims from token")
	}

	userIDClaim, ok := claims["user_id"].(string)
	if !ok {
		log.Println("user_id not found in token claims")
		return "", fmt.Errorf("user_id not found in token claims")
	}

	return userIDClaim, nil
}

func SendMessage(message *models.Message) error {
	db := database.DB

	log.Printf("Saving message to database: %+v\n", message)

	if result := db.Create(&message); result.Error != nil {
		log.Println("Error saving message to database:", result.Error)
		return result.Error
	}

	log.Println("Message saved to database:", message)

	return nil
}

func GetNotifications(c *fiber.Ctx) error {
	db := database.DB
	userID := c.Locals("user_id").(string)

	if userID == "" {
		return helpers.HandleError(c, fiber.StatusUnauthorized, "Unauthorized: missing user_id", nil)
	}

	type NotificationResponse struct {
		Message   string    `json:"message"`
		CreatedAt time.Time `json:"created_at"`
		Category  string    `json:"category"`
	}

	var notifications []NotificationResponse
	if err := db.Table("notifications").
		Select("message, created_at, category").
		Where("user_id = ?", userID).
		Order("created_at desc").
		Scan(&notifications).Error; err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to fetch notifications", err)
	}

	return helpers.HandleSuccess(c, fiber.StatusOK, "Notifications fetched successfully", notifications)
}
