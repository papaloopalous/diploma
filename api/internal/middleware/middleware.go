package middleware

import (
	loggergrpc "api/internal/loggerGRPC"
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
	cookie, err := r.Cookie("authToken")
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, "cookie could not be found", nil)
		loggergrpc.LC.LogInfo("auth", "missing authToken cookie", map[string]string{"details": err.Error()})
		return
	}

	token, err := p.Token.ParseJWT(cookie.Value)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, "invalid token", nil)
		loggergrpc.LC.LogError("auth", "failed to parse JWT", map[string]string{"details": err.Error()})
		return
	}

	userID, role, err := p.Session.GetSession(token.SessionID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusUnauthorized, false, "session could not be found", nil)
		loggergrpc.LC.LogInfo("auth", "session not found", map[string]string{"ID": token.SessionID.String()})
		return
	}

	if role != targetRole && targetRole != "any" {
		response.WriteAPIResponse(w, http.StatusUnauthorized, false, "permission denied", nil)
		loggergrpc.LC.LogInfo("auth", "permission denied", map[string]string{
			"ID":   userID.String(),
			"role": role,
			"need": targetRole,
			"path": r.URL.Path,
		})
		return
	}

	ctx := context.WithValue(r.Context(), userKey, userID)
	next.ServeHTTP(w, r.WithContext(ctx))
}

func (p *MiddlewareHandler) CheckStudent(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { p.CheckSes(w, r, next, "student") })
}

func (p *MiddlewareHandler) CheckTeacher(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { p.CheckSes(w, r, next, "teacher") })
}

func (p *MiddlewareHandler) CheckAny(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { p.CheckSes(w, r, next, "any") })
}

func GetContext(ctx context.Context) (userID uuid.UUID) {
	userID = ctx.Value(userKey).(uuid.UUID)
	return userID
}
