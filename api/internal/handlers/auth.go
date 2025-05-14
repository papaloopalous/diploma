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
)

type AuthHandler struct {
	User    repo.UserRepo
	Token   repo.TokenRepo
	Session repo.SessionRepo
	secret  string
}

var serverSecretKey []byte = []byte("863d268fe1fbedad03c347670de5580d4c44486228c0bb8108840e08b6aea204")

func (p *AuthHandler) EncryptionKey(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ClientPublic string `json:"clientPublic"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, "Invalid request", nil)
		return
	}

	secret, err := encryption.DeriveSharedKeyHex(req.ClientPublic)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, "Invalid public key", nil)
		return
	}

	p.secret = secret

	response.WriteAPIResponse(w, http.StatusOK, true, "", map[string]string{
		"serverPublic": encryption.GetServerPublicKey(),
	})
}

func (p *AuthHandler) LogIN(w http.ResponseWriter, r *http.Request) {
	var requestData map[string]string
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ErrBadRequest, nil)
		loggergrpc.LC.LogError(messages.ServiceAuth, messages.ErrDecodeRequest, map[string]string{messages.LogDetails: err.Error()})
		return
	}

	key := p.secret

	encryptedUsername := requestData[messages.ReqUsername]
	encryptedPassword := requestData[messages.ReqPassword]

	username, err := encryption.DecryptData(encryptedUsername, key)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ErrDecryption, nil)
		loggergrpc.LC.LogError(messages.ServiceAuth, messages.ErrDecrypt, map[string]string{messages.LogDetails: err.Error()})
		return
	}

	password, err := encryption.DecryptData(encryptedPassword, key)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ErrDecryption, nil)
		loggergrpc.LC.LogError(messages.ServiceAuth, messages.ErrDecrypt, map[string]string{messages.LogDetails: err.Error()})
		return
	}

	newPassword, err := encryption.EncryptData(password, string(serverSecretKey))
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ErrEncryption, nil)
		loggergrpc.LC.LogError(messages.ServiceAuth, messages.ErrEncryption, map[string]string{messages.LogDetails: err.Error()})
		return
	}

	userID, userRole, err := p.User.CheckPass(username, newPassword)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusUnauthorized, false, err.Error(), nil)
		loggergrpc.LC.LogInfo(messages.ServiceAuth, messages.ErrAuth, map[string]string{messages.LogDetails: err.Error()})
		return
	}

	sessionID := uuid.New()

	token, err := p.Token.GenerateJWT(sessionID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ErrSessionSet, nil)
		loggergrpc.LC.LogError(messages.ServiceAuth, messages.ErrGenToken, map[string]string{messages.LogDetails: err.Error()})
		return
	}

	err = p.Session.SetSession(sessionID, userID, userRole)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ErrSessionSet, nil)
		loggergrpc.LC.LogError(messages.ServiceAuth, messages.ErrSessionSet, map[string]string{messages.LogDetails: err.Error()})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     messages.CookieAuthToken,
		Value:    token,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		Expires:  time.Now().Add(10 * time.Minute),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     messages.CookieUserRole,
		Value:    userRole,
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		Expires:  time.Now().Add(10 * time.Minute),
	})

	response.WriteAPIResponse(w, http.StatusOK, true, messages.StatusAuth, nil)
	loggergrpc.LC.LogInfo(messages.ServiceAuth, messages.StatusUserAuth, map[string]string{messages.LogUserID: userID.String()})
}

func (p *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var requestData map[string]string
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ErrBadRequest, nil)
		loggergrpc.LC.LogError(messages.ServiceAuth, messages.ErrDecodeRequest, map[string]string{messages.LogDetails: err.Error()})
		return
	}

	key := p.secret

	encryptedUsername := requestData[messages.ReqUsername]
	encryptedPassword := requestData[messages.ReqPassword]
	role := requestData[messages.ReqRole]

	username, err := encryption.DecryptData(encryptedUsername, key)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ErrDecryption, nil)
		loggergrpc.LC.LogError(messages.ServiceAuth, messages.ErrDecrypt, map[string]string{messages.LogDetails: err.Error()})
		return
	}

	password, err := encryption.DecryptData(encryptedPassword, key)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ErrDecryption, nil)
		loggergrpc.LC.LogError(messages.ServiceAuth, messages.ErrDecrypt, map[string]string{messages.LogDetails: err.Error()})
		return
	}

	newPassword, err := encryption.EncryptData(password, string(serverSecretKey))
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ErrEncryption, nil)
		loggergrpc.LC.LogError(messages.ServiceAuth, messages.ErrEncryption, map[string]string{messages.LogDetails: err.Error()})
		return
	}

	userID, err := p.User.CreateAccount(username, newPassword, role)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, err.Error(), nil)
		loggergrpc.LC.LogInfo(messages.ServiceAuth, messages.ErrCeateAcc, map[string]string{messages.LogDetails: err.Error()})
		return
	}

	sessionID := uuid.New()

	token, err := p.Token.GenerateJWT(sessionID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ErrSessionSet, nil)
		loggergrpc.LC.LogError(messages.ServiceAuth, messages.ErrGenToken, map[string]string{messages.LogDetails: err.Error()})
		return
	}

	err = p.Session.SetSession(sessionID, userID, role)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ErrSessionSet, nil)
		loggergrpc.LC.LogError(messages.ServiceAuth, messages.ErrSessionSet, map[string]string{messages.LogDetails: err.Error()})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     messages.CookieAuthToken,
		Value:    token,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		Expires:  time.Now().Add(10 * time.Minute),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     messages.CookieUserRole,
		Value:    role,
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		Expires:  time.Now().Add(10 * time.Minute),
	})

	response.WriteAPIResponse(w, http.StatusOK, true, messages.StatusAuth, nil)
	loggergrpc.LC.LogInfo(messages.ServiceAuth, messages.StatusUserAuth, map[string]string{messages.LogUserID: userID.String()})
}

func (p *AuthHandler) LogOUT(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(messages.CookieAuthToken)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ErrNoToken, nil)
		return
	}

	token, err := p.Token.ParseJWT(cookie.Value)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ErrBadToken, nil)
		loggergrpc.LC.LogError(messages.ServiceAuth, messages.ErrParseToken, map[string]string{messages.LogDetails: err.Error()})
		return
	}

	userID, err := p.Session.DeleteSession(token.SessionID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusNotFound, false, err.Error(), nil)
		loggergrpc.LC.LogError(messages.ServiceAuth, err.Error(), map[string]string{
			messages.LogSessionID: token.SessionID.String(),
		})
		return
	}

	clearCookie := func(name string) {
		http.SetCookie(w, &http.Cookie{
			Name:     name,
			Value:    "",
			HttpOnly: name == messages.CookieAuthToken,
			MaxAge:   -1,
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
		})
	}

	clearCookie(messages.CookieAuthToken)
	clearCookie(messages.CookieUserRole)

	response.WriteAPIResponse(w, http.StatusOK, true, messages.StatusLogOut, nil)
	loggergrpc.LC.LogInfo(messages.ServiceAuth, messages.StatusUserLogOut, map[string]string{messages.LogUserID: userID.String()})
}
