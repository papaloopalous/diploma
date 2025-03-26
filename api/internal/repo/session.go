package repo

import (
	errlist "api/internal/errList"
	"errors"
	"time"

	"github.com/google/uuid"
)

type SessionData struct {
	id        []uuid.UUID
	userID    []uuid.UUID
	role      []string
	expiresAt []time.Time
}

type SessionRepo interface {
	GetSession(sessionID uuid.UUID) (userID uuid.UUID, role string, expiresAt time.Time, err error)
	SetSession(sessionID uuid.UUID, userID uuid.UUID, role string)
	DeleteSession(sessionID uuid.UUID) (userID uuid.UUID, err error)
}

var _ SessionRepo = &SessionData{}

func NewSessionRepo() *SessionData {
	return &SessionData{
		id:        make([]uuid.UUID, 0),
		userID:    make([]uuid.UUID, 0),
		role:      make([]string, 0),
		expiresAt: make([]time.Time, 0),
	}
}

func (p *SessionData) GetSession(sessionID uuid.UUID) (userID uuid.UUID, role string, expiresAt time.Time, err error) {
	found := false

	for i, val := range p.id {
		if val == sessionID {
			userID = p.userID[i]
			role = p.role[i]
			expiresAt = p.expiresAt[i]
			found = true
			break
		}
	}

	if found && expiresAt.Compare(time.Now()) == 1 {
		return userID, role, expiresAt, nil
	} else {
		return userID, role, expiresAt, errors.New(errlist.ErrNoSession)
	}
}

func (p *SessionData) SetSession(sessionID uuid.UUID, userID uuid.UUID, role string) {
	p.id = append(p.id, sessionID)
	p.userID = append(p.userID, userID)
	p.role = append(p.role, role)
	p.expiresAt = append(p.expiresAt, time.Now().Add(10*time.Minute))
}

func (p *SessionData) DeleteSession(sessionID uuid.UUID) (userID uuid.UUID, err error) {
	index := 0
	found := false

	for i, val := range p.id {
		if val == sessionID {
			found = true
			index = i
			userID = p.userID[i]
			break
		}
	}

	if found {
		p.id = append(p.id[:index], p.id[index+1:]...)
		p.role = append(p.role[:index], p.role[index+1:]...)
		p.userID = append(p.userID[:index], p.userID[index+1:]...)
		p.expiresAt = append(p.expiresAt[:index], p.expiresAt[index+1:]...)

		return userID, nil
	}

	return userID, errors.New(errlist.ErrSesNotFound)
}
