package handlers

import (
	"api/internal/encryption"
	loggergrpc "api/internal/loggerGRPC"
	"api/internal/messages"
	"api/internal/repo"
	"api/internal/response"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"
)

// sessionLifetime - время жизни сессии
var sessionLifetime time.Duration

// AuthHandler обрабатывает запросы аутентификации
type AuthHandler struct {
	User    repo.UserRepo    // Репозиторий пользователей
	Token   repo.TokenRepo   // Репозиторий токенов
	Session repo.SessionRepo // Репозиторий сессий
	secret  string           // Секретный ключ для шифрования
}

var serverSecretKey []byte

func init() {
	serverSecretKey = []byte(viper.GetString("crypto.serverSecretKey"))
	coef := viper.GetInt("session.lifetime") // коэффициент времени жизни сессии (в минутах)
	if coef <= 0 {
		coef = 1
	}

	sessionLifetime = time.Duration(coef) * time.Minute
}

// EncryptionKey обменивается ключами для установки защищенного соединения
func (p *AuthHandler) EncryptionKey(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ClientPublic string `json:"clientPublic"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		loggergrpc.LC.LogError(messages.ServiceAuth, messages.LogErrParamsRequest, map[string]string{
			messages.LogDetails: err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrBadRequest, nil)
		return
	}

	secret, err := encryption.DeriveSharedKeyHex(req.ClientPublic)
	if err != nil {
		loggergrpc.LC.LogError(messages.ServiceAuth, messages.LogErrKeyDerivation, map[string]string{
			messages.LogDetails: err.Error(),
			"client_pub":        req.ClientPublic,
		})
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrInvalidPublicKey, nil)
		return
	}

	p.secret = secret

	serverPublic := encryption.GetServerPublicKey()
	loggergrpc.LC.LogInfo(messages.ServiceAuth, messages.LogStatusParamsSent, map[string]string{
		"server_pub": serverPublic,
	})

	response.WriteAPIResponse(w, http.StatusOK, true, messages.StatusSuccess, map[string]string{
		"serverPublic": serverPublic,
	})
}

// LogIN аутентифицирует пользователя
func (p *AuthHandler) LogIN(w http.ResponseWriter, r *http.Request) {
	var requestData map[string]string
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		loggergrpc.LC.LogError(messages.ServiceAuth, messages.LogErrParamsRequest, map[string]string{
			messages.LogDetails: err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrBadRequest, nil)
		return
	}

	key := p.secret

	encryptedUsername := requestData[messages.ReqUsername]
	encryptedPassword := requestData[messages.ReqPassword]

	username, err := encryption.DecryptData(encryptedUsername, key)
	if err != nil {
		loggergrpc.LC.LogError(messages.ServiceAuth, messages.LogErrDecryption, map[string]string{
			messages.LogDetails: err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrDecryption, nil)
		return
	}

	password, err := encryption.DecryptData(encryptedPassword, key)
	if err != nil {
		loggergrpc.LC.LogError(messages.ServiceAuth, messages.LogErrDecryption, map[string]string{
			messages.LogDetails: err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrDecryption, nil)
		return
	}

	newPassword, err := encryption.EncryptData(password, string(serverSecretKey))
	if err != nil {
		loggergrpc.LC.LogError(messages.ServiceAuth, messages.LogErrEncryption, map[string]string{
			messages.LogDetails: err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrEncryption, nil)
		return
	}

	userID, userRole, err := p.User.CheckPass(username, newPassword)
	if err != nil {
		loggergrpc.LC.LogError(messages.ServiceAuth, messages.LogErrAuthFailed, map[string]string{
			messages.LogDetails:  err.Error(),
			messages.LogUsername: username,
		})
		response.WriteAPIResponse(w, http.StatusUnauthorized, false, messages.ClientErrAuth, nil)
		return
	}

	sessionID := uuid.New()
	token, err := p.Token.GenerateJWT(sessionID)
	if err != nil {
		loggergrpc.LC.LogError(messages.ServiceAuth, messages.LogErrSessionInvalid, map[string]string{
			messages.LogSessionID: sessionID.String(),
			messages.LogDetails:   err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrSessionCreation, nil)
		return
	}

	err = p.Session.SetSession(sessionID, userID, userRole, sessionLifetime)
	if err != nil {
		loggergrpc.LC.LogError(messages.ServiceAuth, messages.LogErrSessionInvalid, map[string]string{
			messages.LogSessionID: sessionID.String(),
			messages.LogDetails:   err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrSessionCreation, nil)
		return
	}

	setCookie(w, messages.CookieAuthToken, token, true)
	setCookie(w, messages.CookieUserRole, userRole, false)

	loggergrpc.LC.LogInfo(messages.ServiceAuth, messages.LogStatusUserAuth, map[string]string{
		messages.LogUserID:   userID.String(),
		messages.LogUserRole: userRole,
	})
	response.WriteAPIResponse(w, http.StatusOK, true, messages.StatusAuth, nil)
}

// Register регистрирует нового пользователя
func (p *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var requestData map[string]string
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		loggergrpc.LC.LogError(messages.ServiceAuth, messages.LogErrParamsRequest, map[string]string{
			messages.LogDetails: err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrBadRequest, nil)
		return
	}

	key := p.secret

	encryptedUsername := requestData[messages.ReqUsername]
	encryptedPassword := requestData[messages.ReqPassword]
	role := requestData[messages.ReqRole]

	username, err := encryption.DecryptData(encryptedUsername, key)
	if err != nil {
		loggergrpc.LC.LogError(messages.ServiceAuth, messages.LogErrDecryption, map[string]string{
			messages.LogDetails: err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrDecryption, nil)
		return
	}

	password, err := encryption.DecryptData(encryptedPassword, key)
	if err != nil {
		loggergrpc.LC.LogError(messages.ServiceAuth, messages.LogErrDecryption, map[string]string{
			messages.LogDetails: err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrDecryption, nil)
		return
	}

	newPassword, err := encryption.EncryptData(password, string(serverSecretKey))
	if err != nil {
		loggergrpc.LC.LogError(messages.ServiceAuth, messages.LogErrEncryption, map[string]string{
			messages.LogDetails: err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrEncryption, nil)
		return
	}

	userID, err := p.User.CreateAccount(username, newPassword, role)
	if err != nil {
		loggergrpc.LC.LogError(messages.ServiceAuth, messages.LogErrDBQuery, map[string]string{
			messages.LogDetails:  err.Error(),
			messages.LogUsername: username,
		})
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrCreateAccount, nil)
		return
	}

	if userID == uuid.Nil {
		loggergrpc.LC.LogError(messages.ServiceAuth, messages.LogErrUserExists, map[string]string{
			messages.LogUsername: username,
		})
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrUserExists, nil)
		return
	}

	sessionID := uuid.New()

	token, err := p.Token.GenerateJWT(sessionID)
	if err != nil {
		loggergrpc.LC.LogError(messages.ServiceAuth, messages.LogErrTokenGeneration, map[string]string{
			messages.LogSessionID: sessionID.String(),
			messages.LogDetails:   err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrSessionCreation, nil)
		return
	}

	err = p.Session.SetSession(sessionID, userID, role, sessionLifetime)
	if err != nil {
		loggergrpc.LC.LogError(messages.ServiceAuth, messages.LogErrSessionInvalid, map[string]string{
			messages.LogSessionID: sessionID.String(),
			messages.LogDetails:   err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrSessionCreation, nil)
		return
	}

	setCookie(w, messages.CookieAuthToken, token, true)
	setCookie(w, messages.CookieUserRole, role, false)

	response.WriteAPIResponse(w, http.StatusOK, true, messages.StatusAuth, nil)
	loggergrpc.LC.LogInfo(messages.ServiceAuth, messages.LogStatusUserAuth, map[string]string{messages.LogUserID: userID.String()})
}

// LogOUT завершает сессию пользователя
func (p *AuthHandler) LogOUT(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(messages.CookieAuthToken)
	if err != nil {
		loggergrpc.LC.LogError(messages.ServiceAuth, messages.LogErrSessionInvalid, map[string]string{
			messages.LogDetails: err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusOK, false, messages.ClientErrSessionExpired, nil)
		return
	}

	token, err := p.Token.ParseJWT(cookie.Value)
	if err != nil {
		loggergrpc.LC.LogError(messages.ServiceAuth, messages.LogErrSessionInvalid, map[string]string{
			messages.LogSessionID: token.SessionID.String(),
			messages.LogDetails:   err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusOK, false, messages.ClientErrSessionExpired, nil)
		return
	}

	userID, err := p.Session.DeleteSession(token.SessionID)
	if err != nil {
		loggergrpc.LC.LogError(messages.ServiceAuth, messages.LogErrSessionDelete, map[string]string{
			messages.LogSessionID: token.SessionID.String(),
			messages.LogDetails:   err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusOK, false, messages.ClientErrSessionExpired, nil)
		return
	}

	clearCookie(w, messages.CookieAuthToken)
	clearCookie(w, messages.CookieUserRole)

	response.WriteAPIResponse(w, http.StatusOK, true, messages.StatusLogOut, nil)
	loggergrpc.LC.LogInfo(messages.ServiceAuth, messages.LogStatusUserLogOut, map[string]string{messages.LogUserID: userID.String()})
}

// setCookie устанавливает cookie с заданными параметрами
func setCookie(w http.ResponseWriter, name, value string, httpOnly bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		HttpOnly: httpOnly,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		Expires:  time.Now().Add(sessionLifetime),
	})
}

// clearCookie удаляет cookie
func clearCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		HttpOnly: name == messages.CookieAuthToken,
		MaxAge:   -1,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})
}
