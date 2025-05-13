package middleware

import (
	loggergrpc "api/internal/loggerGRPC"
	"api/internal/messages"
	"api/internal/repo"
	"api/internal/response"
	"context"
	"net/http"

	"github.com/google/uuid"
)

type MiddlewareHandler struct {
	User    repo.UserRepo
	Token   repo.TokenRepo
	Session repo.SessionRepo
}

type contextKey string

const userKey contextKey = "UserKey"

func (p *MiddlewareHandler) CheckSes(w http.ResponseWriter, r *http.Request, next http.Handler, targetRole string) {
	cookie, err := r.Cookie(messages.CookieAuthToken)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ErrNoCookie, nil)
		loggergrpc.LC.LogInfo(messages.ServiceAuth, messages.ErrNoAuthToken, map[string]string{messages.LogDetails: err.Error()})
		return
	}

	token, err := p.Token.ParseJWT(cookie.Value)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ErrBadToken, nil)
		loggergrpc.LC.LogError(messages.ServiceAuth, messages.ErrParseToken, map[string]string{messages.LogDetails: err.Error()})
		return
	}

	userID, role, err := p.Session.GetSession(token.SessionID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusUnauthorized, false, messages.ErrNoSession, nil)
		loggergrpc.LC.LogError(messages.ServiceAuth, messages.ErrSessionNotFound, map[string]string{messages.LogSessionID: token.SessionID.String(),
			messages.LogDetails: err.Error()})
		return
	}

	if role != targetRole && targetRole != "any" {
		response.WriteAPIResponse(w, http.StatusUnauthorized, false, messages.StatusNoPermission, nil)
		loggergrpc.LC.LogInfo(messages.ServiceAuth, messages.StatusUserNoPermission, map[string]string{
			messages.LogUserID:   userID.String(),
			messages.LogUserRole: role,
			messages.LogNeedRole: targetRole,
			messages.LogReqPath:  r.URL.Path,
		})
		return
	}

	ctx := context.WithValue(r.Context(), userKey, userID)
	next.ServeHTTP(w, r.WithContext(ctx))
}

func (p *MiddlewareHandler) CheckStudent(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { p.CheckSes(w, r, next, messages.RoleStudent) })
}

func (p *MiddlewareHandler) CheckTeacher(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { p.CheckSes(w, r, next, messages.RoleTeacher) })
}

func (p *MiddlewareHandler) CheckAny(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { p.CheckSes(w, r, next, "any") })
}

func GetContext(ctx context.Context) (userID uuid.UUID) {
	userID = ctx.Value(userKey).(uuid.UUID)
	return userID
}
