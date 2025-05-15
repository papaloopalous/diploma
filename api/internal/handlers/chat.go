package handlers

import (
	loggergrpc "api/internal/loggerGRPC"
	"api/internal/messages"
	"api/internal/middleware"
	"api/internal/repo"
	"api/internal/response"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type ChatHandler struct {
	User repo.UserRepo
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	conn     *websocket.Conn
	userID   uuid.UUID
	isActive bool
}

type Message struct {
	ID       string    `json:"id"`
	Type     string    `json:"type"`
	RoomID   string    `json:"roomId"`
	SenderID uuid.UUID `json:"senderId"`
	Text     string    `json:"text"`
	SentAt   time.Time `json:"sentAt"`
	IsSender bool      `json:"isSender"`
	Status   string    `json:"status"`
}

type Room struct {
	clients     []*Client
	clientsLock sync.Mutex
	user1ID     uuid.UUID
	user2ID     uuid.UUID
	messages    []Message
}

var rooms = make(map[string]*Room)
var roomsLock sync.Mutex

func findRoomByUsers(user1, user2 uuid.UUID) (string, bool) {
	roomsLock.Lock()
	defer roomsLock.Unlock()

	for roomID, room := range rooms {
		if (room.user1ID == user1 && room.user2ID == user2) ||
			(room.user1ID == user2 && room.user2ID == user1) {
			return roomID, true
		}
	}
	return "", false
}

type ChatMessage struct {
	ID       uuid.UUID `json:"id"`
	RoomID   string    `json:"room_id"`
	SenderID uuid.UUID `json:"sender_id"`
	Text     string    `json:"text"`
	Status   string    `json:"status"`
	SentAt   time.Time `json:"sent_at"`
}

type MessageEvent struct {
	Type    string      `json:"type"`
	Payload ChatMessage `json:"payload"`
}

type CreateRoomRequest struct {
	OtherUserID string `json:"otherUserId"`
}

func (h *ChatHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ErrBadRequest, nil)
		loggergrpc.LC.LogError(messages.ServiceChat, messages.ErrDecodeRequest, map[string]string{
			messages.LogDetails: err.Error(),
		})
		return
	}

	userID := middleware.GetContext(r.Context())
	if userID == uuid.Nil {
		response.WriteAPIResponse(w, http.StatusUnauthorized, false, messages.ErrBadUserID, nil)
		loggergrpc.LC.LogError(messages.ServiceChat, messages.ErrParseUserID, nil)
		return
	}

	otherID, err := uuid.Parse(req.OtherUserID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ErrBadUserID, nil)
		loggergrpc.LC.LogError(messages.ServiceChat, messages.ErrParseUserID, map[string]string{
			messages.LogUserID: req.OtherUserID,
		})
		return
	}

	_, err = h.User.FindUser(userID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusNotFound, false, messages.ErrUserNotFound, nil)
		loggergrpc.LC.LogError(messages.ServiceChat, messages.ErrUserNotFound, map[string]string{
			messages.LogUserID: userID.String(),
		})
		return
	}

	_, err = h.User.FindUser(otherID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusNotFound, false, messages.ErrUserNotFound, nil)
		loggergrpc.LC.LogError(messages.ServiceChat, messages.ErrUserNotFound, map[string]string{
			messages.LogUserID: otherID.String(),
		})
		return
	}

	if existingRoomID, found := findRoomByUsers(userID, otherID); found {
		response.WriteAPIResponse(w, http.StatusOK, true, messages.StatusRoomExists, map[string]string{
			"roomId": existingRoomID,
		})
		loggergrpc.LC.LogInfo(messages.ServiceChat, messages.StatusRoomExists, map[string]string{
			messages.LogRoomID:  existingRoomID,
			messages.LogUserID:  userID.String(),
			messages.LogOtherID: otherID.String(),
		})
		return
	}

	roomID := uuid.New().String()
	roomsLock.Lock()
	rooms[roomID] = &Room{
		clients:  make([]*Client, 0, 2),
		user1ID:  userID,
		user2ID:  otherID,
		messages: make([]Message, 0),
	}
	roomsLock.Unlock()

	response.WriteAPIResponse(w, http.StatusCreated, true, messages.StatusRoomCreated, map[string]string{
		"roomId": roomID,
	})
	loggergrpc.LC.LogInfo(messages.ServiceChat, messages.StatusRoomCreated, map[string]string{
		messages.LogRoomID:  roomID,
		messages.LogUserID:  userID.String(),
		messages.LogOtherID: otherID.String(),
	})
}

func (h *ChatHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("room")
	if roomID == "" {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ErrNoRoomID, nil)
		return
	}

	currentUserID := middleware.GetContext(r.Context())
	if currentUserID == uuid.Nil {
		response.WriteAPIResponse(w, http.StatusUnauthorized, false, messages.ErrBadUserID, nil)
		loggergrpc.LC.LogError(messages.ServiceChat, messages.ErrParseUserID, nil)
		return
	}

	roomsLock.Lock()
	room, exists := rooms[roomID]
	if !exists {
		roomsLock.Unlock()
		response.WriteAPIResponse(w, http.StatusNotFound, false, messages.ErrRoomNotFound, nil)
		loggergrpc.LC.LogError(messages.ServiceChat, messages.ErrRoomNotFound, map[string]string{
			messages.LogRoomID: roomID,
		})
		return
	}

	if room.user1ID != currentUserID && room.user2ID != currentUserID {
		roomsLock.Unlock()
		response.WriteAPIResponse(w, http.StatusForbidden, false, messages.ErrNoRoomAccess, nil)
		loggergrpc.LC.LogError(messages.ServiceChat, messages.ErrNoRoomAccess, map[string]string{
			messages.LogRoomID: roomID,
			messages.LogUserID: currentUserID.String(),
		})
		return
	}
	roomsLock.Unlock()

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		loggergrpc.LC.LogError(messages.ServiceChat, messages.ErrUpgradeConn, map[string]string{
			messages.LogDetails: err.Error(),
		})
		return
	}
	defer conn.Close()

	room.clientsLock.Lock()
	for i, existingClient := range room.clients {
		if existingClient.userID == currentUserID {
			room.clients = append(room.clients[:i], room.clients[i+1:]...)
			break
		}
	}

	activeUsers := make(map[uuid.UUID]bool)
	for _, c := range room.clients {
		if c.isActive {
			activeUsers[c.userID] = true
		}
	}

	if len(activeUsers) >= 2 {
		room.clientsLock.Unlock()
		conn.WriteMessage(websocket.TextMessage, []byte(messages.ErrRoomFull))
		return
	}

	client := &Client{
		conn:     conn,
		userID:   currentUserID,
		isActive: true,
	}

	for i, msg := range room.messages {
		msgCopy := msg
		msgCopy.IsSender = (msg.SenderID == currentUserID)

		if msg.SenderID != currentUserID && msg.Status == "sent" {
			room.messages[i].Status = "delivered"

			statusUpdate := Message{
				ID:     msg.ID,
				Type:   "status",
				Status: "delivered",
			}

			for _, c := range room.clients {
				if msg.SenderID == c.userID {
					err := c.conn.WriteJSON(statusUpdate)
					if err != nil {
						log.Printf("Error sending status update: %v", err)
					}
					break
				}
			}
		}

		err := conn.WriteJSON(msgCopy)
		if err != nil {
			log.Printf("Error sending history: %v", err)
		}
	}

	room.clients = append(room.clients, client)
	room.clientsLock.Unlock()

	loggergrpc.LC.LogInfo(messages.ServiceChat, messages.StatusUserConnected, map[string]string{
		messages.LogRoomID: roomID,
	})

	for {
		var message Message
		err := conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("Unexpected close error: %v", err)
			}
			client.isActive = false
			break
		}

		if message.Type == "message" {
			message.ID = uuid.New().String()
			message.SentAt = time.Now()
			message.SenderID = currentUserID
			message.RoomID = roomID
			message.Status = "sent"
			message.IsSender = true

			conn.WriteJSON(message)

			room.clientsLock.Lock()
			room.messages = append(room.messages, message)

			for _, c := range room.clients {
				if c.userID != currentUserID && c.isActive {
					msgCopy := message
					msgCopy.IsSender = false
					err := c.conn.WriteJSON(msgCopy)
					if err != nil {
						log.Printf("Error sending message: %v", err)
						continue
					}

					message.Status = "delivered"
					err = conn.WriteJSON(Message{
						ID:     message.ID,
						Type:   "status",
						Status: "delivered",
					})
					if err != nil {
						log.Printf("Error sending status update: %v", err)
					}
				}
			}
			room.clientsLock.Unlock()
		}
	}

	room.clientsLock.Lock()
	for i, c := range room.clients {
		if c.conn == conn {
			room.clients = append(room.clients[:i], room.clients[i+1:]...)
			break
		}
	}
	room.clientsLock.Unlock()
}
