package middleware

import (
	errlist "api/internal/errList"
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
		response.APIRespond(w, http.StatusBadRequest, errlist.ErrNoCookie, err.Error(), "ERROR")
		return
	}

	token, err := p.Token.ParseJWT(cookie.Value)
	if err != nil {
		response.APIRespond(w, http.StatusInternalServerError, errlist.ErrTokenParse, err.Error(), "ERROR")
		return
	}

	userID, role, err := p.Session.GetSession(token.SessionID)

	if err != nil {
		response.APIRespond(w, http.StatusUnauthorized, errlist.ErrNoSession, "", "ERROR")
		return
	}

	if role != targetRole {
		response.APIRespond(w, http.StatusUnauthorized, errlist.ErrNoPermission, "id: "+userID.String(), "ERROR")
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

func GetContext(ctx context.Context) (userID uuid.UUID) {
	userID = ctx.Value(userKey).(uuid.UUID)
	return userID
}
