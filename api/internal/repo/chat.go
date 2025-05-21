package repo

import (
	"api/internal/proto/chatpb"
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ChatRepoGRPC реализует взаимодействие с сервисом чата через gRPC
type ChatRepoGRPC struct {
	db chatpb.ChatServiceClient // gRPC клиент для взаимодействия с сервисом чата
}

// Проверка реализации интерфейса
var _ ChatRepo = (*ChatRepoGRPC)(nil)

// NewChatRepo создает новый экземпляр репозитория чата
func NewChatRepo(conn *grpc.ClientConn) *ChatRepoGRPC {
	return &ChatRepoGRPC{
		db: chatpb.NewChatServiceClient(conn),
	}
}

const (
	chatToken = "chat-token"
)

// CreateRoom создает новую комнату чата для двух пользователей
func (r *ChatRepoGRPC) CreateRoom(user1, user2 uuid.UUID) (string, bool, error) {
	md := metadata.New(map[string]string{
		authorization: bearer + chatToken,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	resp, err := r.db.CreateRoom(ctx, &chatpb.CreateRoomRequest{
		User1Id: user1.String(),
		User2Id: user2.String(),
	})
	if err != nil {
		return "", false, err
	}
	return resp.RoomId, resp.AlreadyExists, nil
}

// History возвращает историю сообщений для указанной комнаты
func (r *ChatRepoGRPC) History(roomID string) ([]ChatMessage, error) {
	md := metadata.New(map[string]string{
		authorization: bearer + chatToken,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	resp, err := r.db.History(ctx, &chatpb.RoomIDRequest{RoomId: roomID})
	if err != nil {
		return nil, err
	}

	out := make([]ChatMessage, 0, len(resp.Messages))
	for _, m := range resp.Messages {
		out = append(out, ChatMessage{
			ID:       uuid.MustParse(m.Id),
			RoomID:   m.RoomId,
			SenderID: uuid.MustParse(m.SenderId),
			Text:     m.Text,
			SentAt:   m.SentAt.AsTime(),
			Status:   m.Status,
		})
	}
	return out, nil
}

// SaveMessage сохраняет новое сообщение в базе данных
func (r *ChatRepoGRPC) SaveMessage(msg ChatMessage) error {
	md := metadata.New(map[string]string{
		authorization: bearer + chatToken,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	_, err := r.db.SendMessage(ctx, &chatpb.SendMessageRequest{
		Message: &chatpb.MessageInfo{
			Id:       msg.ID.String(),
			RoomId:   msg.RoomID,
			SenderId: msg.SenderID.String(),
			Text:     msg.Text,
			SentAt:   timestamppb.New(msg.SentAt),
			Status:   msg.Status,
		},
	})
	return err
}

// UpdateStatus обновляет статус сообщения
func (r *ChatRepoGRPC) UpdateStatus(msgID uuid.UUID, status chatpb.MessageStatus) error {
	md := metadata.New(map[string]string{
		authorization: bearer + chatToken,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	_, err := r.db.UpdateStatus(ctx, &chatpb.UpdateStatusRequest{
		Id:     msgID.String(),
		Status: status,
	})
	return err
}
