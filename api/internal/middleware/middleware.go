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

// MiddlewareHandler содержит репозитории для проверки аутентификации и авторизации
type MiddlewareHandler struct {
	User    repo.UserRepo    // Репозиторий пользователей
	Token   repo.TokenRepo   // Репозиторий токенов
	Session repo.SessionRepo // Репозиторий сессий
}

// contextKey определяет тип ключа для контекста
type contextKey string

// userKey - ключ для хранения ID пользователя в контексте
const userKey contextKey = "UserKey"

// CheckSes проверяет сессию и права доступа пользователя
func (p *MiddlewareHandler) CheckSes(w http.ResponseWriter, r *http.Request, next http.Handler, targetRole string) {
	cookie, err := r.Cookie(messages.CookieAuthToken)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrNoCookie, nil)
		loggergrpc.LC.LogInfo(messages.ServiceMiddleware, messages.LogErrNoAuthToken, nil)
		return
	}

	token, err := p.Token.ParseJWT(cookie.Value)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrBadToken, nil)
		loggergrpc.LC.LogError(messages.ServiceMiddleware, messages.LogErrParseToken, map[string]string{
			messages.LogDetails: err.Error(),
		})
		return
	}

	userID, role, err := p.Session.GetSession(token.SessionID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusUnauthorized, false, messages.ClientErrNoSession, nil)
		loggergrpc.LC.LogError(messages.ServiceMiddleware, messages.LogErrSessionNotFound, map[string]string{
			messages.LogSessionID: token.SessionID.String(),
			messages.LogDetails:   err.Error(),
		})
		return
	}

	if role != targetRole && targetRole != "any" {
		response.WriteAPIResponse(w, http.StatusUnauthorized, false, messages.StatusNoPermission, nil)
		loggergrpc.LC.LogInfo(messages.ServiceMiddleware, messages.LogStatusUserNoPermission, map[string]string{
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

// CheckStudent проверяет наличие прав студента
// Оборачивает переданный обработчик проверкой прав доступа для роли "student"
func (p *MiddlewareHandler) CheckStudent(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.CheckSes(w, r, next, messages.RoleStudent)
	})
}

// CheckTeacher проверяет наличие прав преподавателя
// Оборачивает переданный обработчик проверкой прав доступа для роли "teacher"
func (p *MiddlewareHandler) CheckTeacher(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.CheckSes(w, r, next, messages.RoleTeacher)
	})
}

// CheckAny проверяет наличие любой роли пользователя
// Оборачивает переданный обработчик проверкой аутентификации без проверки конкретной роли
func (p *MiddlewareHandler) CheckAny(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.CheckSes(w, r, next, "any")
	})
}

// GetContext извлекает ID пользователя из контекста
func GetContext(ctx context.Context) (userID uuid.UUID) {
	userID = ctx.Value(userKey).(uuid.UUID)
	return userID
}
