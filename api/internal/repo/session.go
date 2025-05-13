package repo

import (
	"api/internal/messages"
	"api/internal/proto/sessionpb"
	"context"
	"errors"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type SessionRepoGRPC struct {
	db sessionpb.SessionServiceClient
}

func NewSessionRepo(conn *grpc.ClientConn) *SessionRepoGRPC {
	return &SessionRepoGRPC{
		db: sessionpb.NewSessionServiceClient(conn),
	}
}

func (r *SessionRepoGRPC) GetSession(sessionID uuid.UUID) (userID uuid.UUID, role string, err error) {
	ctx := context.Background()
	resp, err := r.db.GetSession(ctx, &sessionpb.SessionIDRequest{
		SessionId: sessionID.String(),
	})
	if err != nil {
		return uuid.Nil, "", errors.New(messages.ErrSessionNotFound)
	}

	userID, err = uuid.Parse(resp.UserId)
	if err != nil {
		return uuid.Nil, "", err
	}

	return userID, resp.Role, nil
}

func (r *SessionRepoGRPC) SetSession(sessionID uuid.UUID, userID uuid.UUID, role string) error {
	ctx := context.Background()
	_, err := r.db.SetSession(ctx, &sessionpb.SetSessionRequest{
		SessionId: sessionID.String(),
		UserId:    userID.String(),
		Role:      role,
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *SessionRepoGRPC) DeleteSession(sessionID uuid.UUID) (userID uuid.UUID, err error) {
	ctx := context.Background()
	resp, err := r.db.DeleteSession(ctx, &sessionpb.SessionIDRequest{
		SessionId: sessionID.String(),
	})
	if err != nil {
		return uuid.Nil, errors.New(messages.ErrSessionNotFound)
	}

	return uuid.Parse(resp.UserId)
}
