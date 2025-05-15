package handlers

import (
	loggergrpc "api/internal/loggerGRPC"
	"api/internal/messages"
	"api/internal/middleware"
	"api/internal/repo"
	"api/internal/response"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	chatpb "api/internal/proto/chatpb"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type ChatHandler struct {
	User repo.UserRepo
	Chat repo.ChatRepo
}

type Client struct {
	conn     *websocket.Conn
	userID   uuid.UUID
	isActive bool
}

type Room struct {
	clients     []*Client
	clientsLock sync.Mutex
	user1ID     uuid.UUID
	user2ID     uuid.UUID
}

var (
	upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	rooms    = make(map[string]*Room)
	roomsMu  sync.Mutex
)

type wsMessage struct {
	ID       string    `json:"id"`
	Type     string    `json:"type"`
	RoomID   string    `json:"roomId"`
	SenderID uuid.UUID `json:"senderId"`
	Text     string    `json:"text"`
	SentAt   time.Time `json:"sentAt"`
	IsSender bool      `json:"isSender"`
	Status   string    `json:"status"`
}

func enumToClient(s chatpb.MessageStatus) string {
	switch s {
	case chatpb.MessageStatus_SENT:
		return "sent"
	case chatpb.MessageStatus_DELIVERED:
		return "delivered"
	case chatpb.MessageStatus_READ:
		return "read"
	default:
		return "sent"
	}
}

type createRoomRequest struct {
	OtherUserID string `json:"otherUserId"`
}

func (h *ChatHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req createRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ErrDecodeRequest, nil)
		loggergrpc.LC.LogError(messages.ServiceChat, messages.ErrDecodeRequest, map[string]string{messages.LogDetails: err.Error()})
		return
	}

	userID := middleware.GetContext(r.Context())
	otherID, err := uuid.Parse(req.OtherUserID)
	if err != nil || userID == uuid.Nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ErrBadUserID, nil)
		return
	}

	if _, err = h.User.FindUser(userID); err != nil {
		response.WriteAPIResponse(w, http.StatusNotFound, false, messages.ErrUserNotFound, nil)
		return
	}
	if _, err = h.User.FindUser(otherID); err != nil {
		response.WriteAPIResponse(w, http.StatusNotFound, false, messages.ErrUserNotFound, nil)
		return
	}

	roomID, existed, err := h.Chat.CreateRoom(userID, otherID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, err.Error(), nil)
		loggergrpc.LC.LogError(messages.ServiceChat, err.Error(), nil)
		return
	}

	code := http.StatusCreated
	msg := messages.StatusRoomCreated
	if existed {
		code = http.StatusOK
		msg = messages.StatusRoomExists
	}

	response.WriteAPIResponse(w, code, true, msg, map[string]string{"roomId": roomID})
	loggergrpc.LC.LogInfo(messages.ServiceChat, msg, map[string]string{
		messages.LogRoomID: roomID,
		messages.LogUserID: userID.String(), messages.LogOtherID: otherID.String(),
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
		return
	}

	history, err := h.Chat.History(roomID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusForbidden, false, messages.ErrNoRoomAccess, nil)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		loggergrpc.LC.LogError(messages.ServiceChat, messages.ErrUpgradeConn,
			map[string]string{messages.LogDetails: err.Error()})
		return
	}
	defer conn.Close()

	roomsMu.Lock()
	room := rooms[roomID]
	if room == nil {
		var u1, u2 uuid.UUID
		if len(history) != 0 {
			u1 = history[0].SenderID
			for _, m := range history {
				if m.SenderID != u1 {
					u2 = m.SenderID
					break
				}
			}
		}
		room = &Room{user1ID: u1, user2ID: u2}
		rooms[roomID] = room
	}
	roomsMu.Unlock()

	room.clientsLock.Lock()
	for i, c := range room.clients {
		if c.userID == currentUserID {
			room.clients = append(room.clients[:i], room.clients[i+1:]...)
			break
		}
	}
	client := &Client{conn: conn, userID: currentUserID, isActive: true}
	room.clients = append(room.clients, client)
	room.clientsLock.Unlock()

	loggergrpc.LC.LogInfo(messages.ServiceChat, messages.StatusUserConnected,
		map[string]string{messages.LogRoomID: roomID, messages.LogUserID: currentUserID.String()})

	for i, m := range history {
		if m.Status == chatpb.MessageStatus_SENT && m.SenderID != currentUserID {
			// в БД
			_ = h.Chat.UpdateStatus(m.ID, chatpb.MessageStatus_DELIVERED)
			history[i].Status = chatpb.MessageStatus_DELIVERED

			room.clientsLock.Lock()
			for _, c := range room.clients {
				if c.userID == m.SenderID && c.isActive {
					_ = c.conn.WriteJSON(wsMessage{
						ID:     m.ID.String(),
						Type:   "status",
						Status: "delivered",
					})
					break
				}
			}
			room.clientsLock.Unlock()
		}
	}

	for _, m := range history {
		_ = conn.WriteJSON(wsMessage{
			ID:       m.ID.String(),
			Type:     "message",
			RoomID:   roomID,
			SenderID: m.SenderID,
			Text:     m.Text,
			SentAt:   m.SentAt,
			IsSender: m.SenderID == currentUserID,
			Status:   enumToClient(m.Status),
		})
	}

	for {
		var incoming wsMessage
		if err := conn.ReadJSON(&incoming); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				loggergrpc.LC.LogError(messages.ServiceChat, err.Error(), nil)
			}
			client.isActive = false
			break
		}

		if incoming.Type != "message" {
			continue
		}

		newMsg := repo.ChatMessage{
			ID:       uuid.New(),
			RoomID:   roomID,
			SenderID: currentUserID,
			Text:     incoming.Text,
			SentAt:   time.Now(),
			Status:   chatpb.MessageStatus_SENT,
		}

		if err := h.Chat.SaveMessage(newMsg); err != nil {
			loggergrpc.LC.LogError(messages.ServiceChat, err.Error(), nil)
			continue
		}

		_ = conn.WriteJSON(wsMessage{
			ID:       newMsg.ID.String(),
			Type:     "message",
			RoomID:   roomID,
			SenderID: currentUserID,
			Text:     newMsg.Text,
			SentAt:   newMsg.SentAt,
			IsSender: true,
			Status:   "sent",
		})

		room.clientsLock.Lock()
		for _, c := range room.clients {
			if c.userID == currentUserID || !c.isActive {
				continue
			}
			if err := c.conn.WriteJSON(wsMessage{
				ID:       newMsg.ID.String(),
				Type:     "message",
				RoomID:   roomID,
				SenderID: currentUserID,
				Text:     newMsg.Text,
				SentAt:   newMsg.SentAt,
				IsSender: false,
				Status:   "sent",
			}); err == nil {
				_ = h.Chat.UpdateStatus(newMsg.ID, chatpb.MessageStatus_DELIVERED)
				_ = conn.WriteJSON(wsMessage{
					ID:     newMsg.ID.String(),
					Type:   "status",
					Status: "delivered",
				})
			}
		}
		room.clientsLock.Unlock()
	}

	room.clientsLock.Lock()
	for i, c := range room.clients {
		if c == client {
			room.clients = append(room.clients[:i], room.clients[i+1:]...)
			break
		}
	}
	room.clientsLock.Unlock()
}
