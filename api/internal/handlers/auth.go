package handlers

import (
	"api/internal/encryption"
	loggergrpc "api/internal/loggerGRPC"
	"api/internal/repo"
	"api/internal/response"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type AuthHandler struct {
	User    repo.UserRepo
	Token   repo.TokenRepo
	Session repo.SessionRepo
}

func (p *AuthHandler) LogIN(w http.ResponseWriter, r *http.Request) {
	var requestData map[string]string
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, "failed to decode the request body", nil)
		loggergrpc.LC.LogError("auth", "failed to decode the request body", map[string]string{"details": err.Error()})
		return
	}

	key, err := encryption.GetEncryptionKey()
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, "decryption error", nil)
		loggergrpc.LC.LogError("auth", "failed to get an encryption key", map[string]string{"details": err.Error()})
		return
	}

	encryptedUsername := requestData["username"]
	encryptedPassword := requestData["password"]

	username, err := encryption.DecryptData(encryptedUsername, key)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, "decryption error", nil)
		loggergrpc.LC.LogError("auth", "failed to decrypt data", map[string]string{"details": err.Error()})
		return
	}

	userID, userRole, err := p.User.CheckPass(username, encryptedPassword)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusUnauthorized, false, err.Error(), nil)
		loggergrpc.LC.LogInfo("auth", "failed to authorize", map[string]string{"details": err.Error()})
		return
	}

	sessionID := uuid.New()

	token, err := p.Token.GenerateJWT(sessionID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, "failed to create a session", nil)
		loggergrpc.LC.LogError("auth", "failed to generate a token", map[string]string{"details": err.Error()})
		return
	}

	p.Session.SetSession(sessionID, userID, userRole)

	http.SetCookie(w, &http.Cookie{
		Name:     "authToken",
		Value:    token,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "userRole",
		Value:    userRole,
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	response.WriteAPIResponse(w, http.StatusOK, true, "authorized", nil)
	loggergrpc.LC.LogInfo("auth", "user authorized", map[string]string{"ID": userID.String()})
}

func (p *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var requestData map[string]string
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, "failed to decode the request body", nil)
		loggergrpc.LC.LogError("auth", "failed to decode the request body", map[string]string{"details": err.Error()})
		return
	}

	key, err := encryption.GetEncryptionKey()
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, "decryption error", nil)
		loggergrpc.LC.LogError("auth", "failed to get an encryption key", map[string]string{"details": err.Error()})
		return
	}

	encryptedUsername := requestData["username"]
	encryptedPassword := requestData["password"]
	role := requestData["role"]

	username, err := encryption.DecryptData(encryptedUsername, key)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, "decryption error", nil)
		loggergrpc.LC.LogError("auth", "failed to decrypt data", map[string]string{"details": err.Error()})
		return
	}

	userID, err := p.User.CreateAccount(username, encryptedPassword, role)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, err.Error(), nil)
		loggergrpc.LC.LogInfo("auth", "failed to create an account", map[string]string{"details": err.Error()})
		return
	}

	sessionID := uuid.New()

	token, err := p.Token.GenerateJWT(sessionID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, "failed to create a session", nil)
		loggergrpc.LC.LogError("auth", "failed to generate a token", map[string]string{"details": err.Error()})
		return
	}

	p.Session.SetSession(sessionID, userID, role)

	http.SetCookie(w, &http.Cookie{
		Name:     "authToken",
		Value:    token,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "userRole",
		Value:    role,
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	response.WriteAPIResponse(w, http.StatusOK, true, "authorized", nil)
	loggergrpc.LC.LogInfo("auth", "user authorized", map[string]string{"ID": userID.String()})
}

func (p *AuthHandler) LogOUT(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("authToken")
	if err != nil {
		response.WriteAPIResponse(w, http.StatusUnauthorized, false, "auth token is missing", nil)
		return
	}

	token, err := p.Token.ParseJWT(cookie.Value)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, "invalid token", nil)
		loggergrpc.LC.LogError("auth", "failed to parse JWT token", map[string]string{"details": err.Error()})
		return
	}

	userID, err := p.Session.DeleteSession(token.SessionID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusNotFound, false, "session could not be found", nil)
		loggergrpc.LC.LogError("auth", "failed to delete session", map[string]string{
			"ID":      token.SessionID.String(),
			"details": err.Error(),
		})
		return
	}

	clearCookie := func(name string) {
		http.SetCookie(w, &http.Cookie{
			Name:     name,
			Value:    "",
			HttpOnly: name == "authToken",
			MaxAge:   -1,
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
		})
	}

	clearCookie("authToken")
	clearCookie("userRole")

	response.WriteAPIResponse(w, http.StatusOK, true, "logged out", nil)
	loggergrpc.LC.LogInfo("auth", "user logged out", map[string]string{"ID": userID.String()})
}
