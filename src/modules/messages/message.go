package messages

import (
	"Backend/src/core/database"
	"Backend/src/core/helpers"
	"Backend/src/core/models"
	"bufio"
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/valyala/fasthttp"
)

// WebSocket upgrader for upgrading HTTP requests to WebSocket
var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

// A map to store WebSocket connections for each community (chat room)
var communityConnections = make(map[string][]*websocket.Conn)
var mu sync.Mutex

// Custom adapter to wrap fasthttp.Response and make it compatible with http.ResponseWriter
type ResponseAdapter struct {
    Response *fiber.Response
}

func (r *ResponseAdapter) Header() http.Header {
	headers := make(http.Header)
	r.Response.Header.VisitAll(func(key, value []byte) {
		headers[string(key)] = append(headers[string(key)], string(value))
	})
	return headers
}

func (r *ResponseAdapter) Write(p []byte) (n int, err error) {
    buf := bufio.NewWriter(r.Response.BodyWriter())
    n, err = buf.Write(p) 
    if err != nil {
        return n, err
    }
    err = buf.Flush()  
    return n, err
}


func (r *ResponseAdapter) WriteHeader(statusCode int) {
	r.Response.SetStatusCode(statusCode)
}

type RequestAdapter struct {
    Request *fasthttp.Request
}

func (r *RequestAdapter) Method() string {
    return string(r.Request.Header.Method()) // Convert []byte to string
}

func (r *RequestAdapter) URL() *url.URL {
    parsedURL := "http://" + string(r.Request.Header.Peek("Host")) + string(r.Request.URI().Path())
    u, err := url.Parse(parsedURL)
    if err != nil {
        return nil
    }
    return u
}

func (r *RequestAdapter) Header() http.Header {
    headers := make(http.Header)
    r.Request.Header.VisitAll(func(key, value []byte) {
        headers[string(key)] = append(headers[string(key)], string(value))
    })
    return headers
}

func (r *RequestAdapter) Body() []byte {
    return r.Request.Body()
}

func WebSocketHandler(c *fiber.Ctx) error {
    req := c.Request()
    res := c.Response()

    // Manually create an *http.Request using the fasthttp request
    httpReq := &http.Request{
        Method:     string(req.Header.Method()),
        Header:     make(http.Header),
        Body:       ioutil.NopCloser(bytes.NewReader(req.Body())),
        RemoteAddr: c.IP(),
    }

    // Copy headers from fasthttp.Request to *http.Request
    req.Header.VisitAll(func(key, value []byte) {
        httpReq.Header.Add(string(key), string(value))
    })

    // Create an adapter for the fasthttp.Response (this will be used as http.ResponseWriter)
    httpRes := &ResponseAdapter{
        Response: res,
    }

    // WebSocket upgrader
    upgrader := websocket.Upgrader{
        CheckOrigin: func(r *http.Request) bool {
            return true // Allow all origins, adjust as needed
        },
    }

    // Upgrade the connection to WebSocket
    conn, err := upgrader.Upgrade(httpRes, httpReq, nil)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).SendString("Failed to upgrade connection")
    }
    defer conn.Close()

    // Store the WebSocket connection in the map for the specific community
    communityID := c.Params("id") // Retrieve community ID from URL params
    mu.Lock()
    communityConnections[communityID] = append(communityConnections[communityID], conn)
    mu.Unlock()

    // Handle WebSocket communication
    for {
        msgType, p, err := conn.ReadMessage()
        if err != nil {
            log.Println("Error reading message:", err)
            break
        }

        // Send the received message back to all connected clients in the same community
        mu.Lock()
        for _, otherConn := range communityConnections[communityID] {
            err := otherConn.WriteMessage(msgType, p)
            if err != nil {
                log.Println("Error sending message:", err)
                continue
            }
        }
        mu.Unlock()
    }

    return nil
}

func SendMessage(c *fiber.Ctx) error {
    db := database.DB

    userIDStr, ok := c.Locals("user_id").(string)
    if !ok || userIDStr == "" {
        return helpers.HandleError(c, fiber.StatusUnauthorized, "Invalid or missing user_id", nil)
    }

    userID, err := uuid.Parse(userIDStr)
    if err != nil {
        return helpers.HandleError(c, fiber.StatusUnauthorized, "Invalid user ID format", err)
    }

    communityIDStr := c.Params("id")
    communityID, err := strconv.Atoi(communityIDStr)
    if err != nil {
        return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid community ID", err)
    }

    message := new(models.Message)

    if err := c.BodyParser(message); err != nil {
        return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid input data", err)
    }

    message.UserID = userID
    message.CommunityID = communityID
    message.CreatedAt = time.Now()

    if result := db.Create(&message); result.Error != nil {
        return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to send message", result.Error)
    }

    mu.Lock()
    for _, conn := range communityConnections[communityIDStr] {
        err := conn.WriteJSON(message)
        if err != nil {
            log.Println("Error sending message:", err)
            continue
        }
    }
    mu.Unlock()

    return helpers.HandleSuccess(c, fiber.StatusCreated, "Message sent successfully", message)
}
