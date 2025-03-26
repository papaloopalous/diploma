package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Структура комнаты
type Room struct {
	clients     []*websocket.Conn
	clientsLock sync.Mutex
}

// Карта активных чатов (roomID → *Room)
var rooms = make(map[string]*Room)
var roomsLock sync.Mutex

func handleConnection(w http.ResponseWriter, r *http.Request) {
	// Получаем ID комнаты из URL (пример: ws://localhost:8080/ws?room=123)
	roomID := r.URL.Query().Get("room")
	if roomID == "" {
		http.Error(w, "Отсутствует room ID", http.StatusBadRequest)
		return
	}

	// Обновляем соединение до WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Ошибка при обновлении соединения:", err)
		return
	}
	defer conn.Close()

	// Добавляем пользователя в комнату
	roomsLock.Lock()
	if _, exists := rooms[roomID]; !exists {
		rooms[roomID] = &Room{}
	}
	room := rooms[roomID]
	roomsLock.Unlock()

	room.clientsLock.Lock()
	if len(room.clients) >= 2 {
		conn.WriteMessage(websocket.TextMessage, []byte("Чат уже заполнен."))
		room.clientsLock.Unlock()
		return
	}
	room.clients = append(room.clients, conn)
	room.clientsLock.Unlock()

	log.Printf("Пользователь подключен к комнате %s\n", roomID)

	// Читаем сообщения от клиента и пересылаем другому пользователю в той же комнате
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Пользователь покинул комнату %s\n", roomID)
			break
		}

		log.Printf("[%s] Сообщение: %s\n", roomID, msg)

		// Отправляем сообщение другому пользователю в этой комнате
		room.clientsLock.Lock()
		for _, c := range room.clients {
			if c != conn {
				err := c.WriteMessage(websocket.TextMessage, msg)
				if err != nil {
					log.Println("Ошибка отправки:", err)
				}
			}
		}
		room.clientsLock.Unlock()
	}

	// Удаляем клиента из комнаты при отключении
	room.clientsLock.Lock()
	for i, c := range room.clients {
		if c == conn {
			room.clients = append(room.clients[:i], room.clients[i+1:]...)
			break
		}
	}
	room.clientsLock.Unlock()
}

func main() {
	http.HandleFunc("/ws", handleConnection)

	port := "8080"
	log.Println("WebSocket сервер запущен на порту", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}
