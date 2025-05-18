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

// ChatHandler обрабатывает запросы чата
type ChatHandler struct {
	User repo.UserRepo // Репозиторий пользователей
	Chat repo.ChatRepo // Репозиторий сообщений чата
}

// Client представляет подключенного пользователя
type Client struct {
	conn     *websocket.Conn // WebSocket соединение
	userID   uuid.UUID       // Идентификатор пользователя
	isActive bool            // Статус подключения
}

// Room представляет комнату чата
type Room struct {
	clients     []*Client  // Подключенные клиенты
	clientsLock sync.Mutex // Мьютекс для безопасного доступа к клиентам
	user1ID     uuid.UUID  // Идентификатор первого пользователя
	user2ID     uuid.UUID  // Идентификатор второго пользователя
}

// Глобальные переменные для работы с WebSocket
var (
	upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	rooms    = make(map[string]*Room) // Карта активных комнат
	roomsMu  sync.Mutex               // Мьютекс для безопасного доступа к комнатам
)

// wsMessage представляет структуру сообщения WebSocket
type wsMessage struct {
	ID       string    `json:"id"`       // Идентификатор сообщения
	Type     string    `json:"type"`     // Тип сообщения
	RoomID   string    `json:"roomId"`   // Идентификатор комнаты
	SenderID uuid.UUID `json:"senderId"` // Идентификатор отправителя
	Text     string    `json:"text"`     // Текст сообщения
	SentAt   time.Time `json:"sentAt"`   // Время отправки
	IsSender bool      `json:"isSender"` // Признак отправителя
	Status   string    `json:"status"`   // Статус сообщения
}

func enumToClient(s chatpb.MessageStatus) string {
	switch s {
	case chatpb.MessageStatus_SENT:
		return messages.ChatStatusSent
	case chatpb.MessageStatus_DELIVERED:
		return messages.ChatStatusDelivered
	case chatpb.MessageStatus_READ:
		return messages.ChatStatusRead
	default:
		return messages.ChatStatusSent
	}
}

type createRoomRequest struct {
	OtherUserID string `json:"otherUserId"`
}

// CreateRoom создает новую комнату для чата между двумя пользователями
func (h *ChatHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req createRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		loggergrpc.LC.LogError(messages.ServiceChat, messages.LogErrDecodeRequest, map[string]string{
			messages.LogDetails: err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrBadRequest, nil)
		return
	}

	userID := middleware.GetContext(r.Context())
	otherID, err := uuid.Parse(req.OtherUserID)
	if err != nil || userID == uuid.Nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrBadID, nil)
		return
	}

	if _, err = h.User.FindUser(userID); err != nil {
		response.WriteAPIResponse(w, http.StatusNotFound, false, messages.ClientErrUserNotFound, nil)
		return
	}
	if _, err = h.User.FindUser(otherID); err != nil {
		response.WriteAPIResponse(w, http.StatusNotFound, false, messages.ClientErrUserNotFound, nil)
		return
	}

	roomID, existed, err := h.Chat.CreateRoom(userID, otherID)
	if err != nil {
		loggergrpc.LC.LogError(messages.ServiceChat, messages.LogErrChatRoomCreate, map[string]string{
			messages.LogDetails: err.Error(),
			messages.LogUserID:  userID.String(),
			messages.LogOtherID: otherID.String(),
		})
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrChatConnect, nil)
		return
	}

	code := http.StatusCreated
	msg := messages.StatusRoomCreated
	if existed {
		code = http.StatusOK
		msg = messages.StatusRoomExists
	}

	response.WriteAPIResponse(w, code, true, msg, map[string]string{messages.LogRoomID: roomID})
	loggergrpc.LC.LogInfo(messages.ServiceChat, messages.LogStatusChatRoomCreated, map[string]string{
		messages.LogRoomID:  roomID,
		messages.LogUserID:  userID.String(),
		messages.LogOtherID: otherID.String(),
	})
}

// HandleConnection обрабатывает входящее WebSocket-соединение
func (h *ChatHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get(messages.ReqRoom)
	if roomID == "" {
		loggergrpc.LC.LogError(messages.ServiceChat, messages.LogErrNoRoomID, nil)
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrNoRoomID, nil)
		return
	}

	currentUserID := middleware.GetContext(r.Context())
	if currentUserID == uuid.Nil {
		response.WriteAPIResponse(w, http.StatusUnauthorized, false, messages.ClientErrBadID, nil)
		return
	}

	history, err := h.Chat.History(roomID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusForbidden, false, messages.ClientErrNoRoomAccess, nil)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		loggergrpc.LC.LogError(messages.ServiceChat, messages.LogErrUpgradeConn, map[string]string{
			messages.LogDetails: err.Error(),
			messages.LogRoomID:  roomID,
		})
		return
	}
	defer conn.Close()

	// Создание или получение комнаты
	roomsMu.Lock()
	room := rooms[roomID]
	if room == nil {
		loggergrpc.LC.LogInfo(messages.ServiceChat, messages.LogStatusRoomCreating, map[string]string{
			messages.LogRoomID: roomID,
		})
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

	// Добавление клиента в комнату
	room.clientsLock.Lock()
	for i, c := range room.clients {
		if c.userID == currentUserID {
			loggergrpc.LC.LogInfo(messages.ServiceChat, messages.LogStatusUserReconnected, map[string]string{
				messages.LogRoomID: roomID,
				messages.LogUserID: currentUserID.String(),
			})
			room.clients = append(room.clients[:i], room.clients[i+1:]...)
			break
		}
	}
	client := &Client{conn: conn, userID: currentUserID, isActive: true}
	room.clients = append(room.clients, client)
	room.clientsLock.Unlock()

	loggergrpc.LC.LogInfo(messages.ServiceChat, messages.StatusUserConnected, map[string]string{
		messages.LogRoomID: roomID,
		messages.LogUserID: currentUserID.String(),
	})

	// Обработка истории сообщений
	for i, m := range history {
		if m.Status == chatpb.MessageStatus_SENT && m.SenderID != currentUserID {
			loggergrpc.LC.LogInfo(messages.ServiceChat, messages.LogStatusMessageDelivered, map[string]string{
				messages.LogRoomID:    roomID,
				messages.LogUserID:    currentUserID.String(),
				messages.LogMessageID: m.ID.String(),
			})
			_ = h.Chat.UpdateStatus(m.ID, chatpb.MessageStatus_DELIVERED)
			history[i].Status = chatpb.MessageStatus_DELIVERED

			room.clientsLock.Lock()
			for _, c := range room.clients {
				if c.userID == m.SenderID && c.isActive {
					_ = c.conn.WriteJSON(wsMessage{
						ID:     m.ID.String(),
						Type:   messages.ChatTypeStatus,
						Status: messages.ChatStatusDelivered,
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
			Type:     messages.ChatTypeMessage,
			RoomID:   roomID,
			SenderID: m.SenderID,
			Text:     m.Text,
			SentAt:   m.SentAt,
			IsSender: m.SenderID == currentUserID,
			Status:   enumToClient(m.Status),
		})
	}

	// Основной цикл обработки сообщений
	for {
		var incoming wsMessage
		if err := conn.ReadJSON(&incoming); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				loggergrpc.LC.LogError(messages.ServiceChat, messages.LogErrWSRead, map[string]string{
					messages.LogDetails: err.Error(),
					messages.LogRoomID:  roomID,
					messages.LogUserID:  currentUserID.String(),
				})
			}
			client.isActive = false
			break
		}

		if incoming.Type != messages.ChatTypeMessage {
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
			loggergrpc.LC.LogError(messages.ServiceChat, messages.LogErrSaveMessage, map[string]string{
				messages.LogDetails: err.Error(),
				messages.LogRoomID:  roomID,
				messages.LogUserID:  currentUserID.String(),
			})
			continue
		}

		_ = conn.WriteJSON(wsMessage{
			ID:       newMsg.ID.String(),
			Type:     messages.ChatTypeMessage,
			RoomID:   roomID,
			SenderID: currentUserID,
			Text:     newMsg.Text,
			SentAt:   newMsg.SentAt,
			IsSender: true,
			Status:   messages.ChatStatusSent,
		})

		loggergrpc.LC.LogInfo(messages.ServiceChat, messages.LogStatusMessageSent, map[string]string{
			messages.LogRoomID:    roomID,
			messages.LogUserID:    currentUserID.String(),
			messages.LogMessageID: newMsg.ID.String(),
		})

		// Отправка сообщения другим клиентам
		room.clientsLock.Lock()
		for _, c := range room.clients {
			if c.userID == currentUserID || !c.isActive {
				continue
			}
			if err := c.conn.WriteJSON(wsMessage{
				ID:       newMsg.ID.String(),
				Type:     messages.ChatTypeMessage,
				RoomID:   roomID,
				SenderID: currentUserID,
				Text:     newMsg.Text,
				SentAt:   newMsg.SentAt,
				IsSender: false,
				Status:   messages.ChatStatusSent,
			}); err == nil {
				loggergrpc.LC.LogInfo(messages.ServiceChat, messages.LogStatusMessageDelivered, map[string]string{
					messages.LogRoomID:     roomID,
					messages.LogUserID:     currentUserID.String(),
					messages.LogReceiverID: c.userID.String(),
					messages.LogMessageID:  newMsg.ID.String(),
				})
			} else {
				loggergrpc.LC.LogError(messages.ServiceChat, messages.LogErrWSSend, map[string]string{
					messages.LogDetails:    err.Error(),
					messages.LogRoomID:     roomID,
					messages.LogUserID:     currentUserID.String(),
					messages.LogReceiverID: c.userID.String(),
				})
			}
		}
		room.clientsLock.Unlock()
	}

	// Удаление клиента при отключении
	room.clientsLock.Lock()
	for i, c := range room.clients {
		if c == client {
			room.clients = append(room.clients[:i], room.clients[i+1:]...)
			loggergrpc.LC.LogInfo(messages.ServiceChat, messages.LogStatusUserDisconnected, map[string]string{
				messages.LogRoomID: roomID,
				messages.LogUserID: currentUserID.String(),
			})
			break
		}
	}
	room.clientsLock.Unlock()
}
