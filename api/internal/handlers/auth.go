package handlers

import (
	"api/internal/encryption"
	errlist "api/internal/errList"
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
		response.APIRespond(w, http.StatusBadRequest, "invalid request", "", "ERROR")
		return
	}

	key, err := encryption.GetEncryptionKey()
	if err != nil {
		response.APIRespond(w, http.StatusInternalServerError, err.Error(), "", "ERROR")
		return
	}

	encryptedUsername := requestData["username"]
	encryptedPassword := requestData["password"]

	username, err := encryption.DecryptData(encryptedUsername, key)
	if err != nil {
		response.APIRespond(w, http.StatusInternalServerError, err.Error(), "", "ERROR")
		return
	}

	userID, userRole, err := p.User.CheckPass(username, encryptedPassword)
	if err != nil {
		response.APIRespond(w, http.StatusUnauthorized, "failed to authenticate", err.Error(), "INFO")
		return
	}

	sessionID := uuid.New()

	token, err := p.Token.GenerateJWT(sessionID)
	if err != nil {
		response.APIRespond(w, http.StatusInternalServerError, "failed to generate a token", err.Error(), "ERROR")
		return
	}

	p.Session.SetSession(sessionID, userID, userRole)

	http.SetCookie(w, &http.Cookie{
		Name:     "authToken",
		Value:    token,
		HttpOnly: true,
	})

	response.APIRespond(w, http.StatusOK, "user authenticated", "id:"+userID.String(), "INFO")
}

func (p *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var requestData map[string]string
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		response.APIRespond(w, http.StatusBadRequest, "invalid request", err.Error(), "ERROR")
		return
	}

	key, err := encryption.GetEncryptionKey()
	if err != nil {
		response.APIRespond(w, http.StatusInternalServerError, err.Error(), "", "ERROR")
		return
	}

	encryptedUsername := requestData["username"]
	encryptedPassword := requestData["password"]
	role := requestData["role"]

	username, err := encryption.DecryptData(encryptedUsername, key)
	if err != nil {
		response.APIRespond(w, http.StatusInternalServerError, err.Error(), "", "ERROR")
		return
	}

	userID, err := p.User.CreateAccount(username, encryptedPassword, role)
	if err != nil {
		response.APIRespond(w, http.StatusBadRequest, "failed to create an account", err.Error(), "INFO")
		return
	}

	sessionID := uuid.New()

	token, err := p.Token.GenerateJWT(sessionID)
	if err != nil {
		response.APIRespond(w, http.StatusInternalServerError, "failed to generate a token", err.Error(), "ERROR")
		return
	}

	p.Session.SetSession(sessionID, userID, role)

	http.SetCookie(w, &http.Cookie{
		Name:     "authToken",
		Value:    token,
		HttpOnly: true,
	})

	response.APIRespond(w, http.StatusCreated, "user authenticated", "id:"+userID.String(), "INFO")
}

func (p *AuthHandler) LogOUT(w http.ResponseWriter, r *http.Request) {
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

	userID, err := p.Session.DeleteSession(token.SessionID)
	if err != nil {
		response.APIRespond(w, http.StatusInternalServerError, errlist.ErrSesDelete, err.Error(), "ERROR")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "authToken",
		Value:    "",
		HttpOnly: true,
		MaxAge:   -1,
	})

	response.APIRespond(w, http.StatusOK, "user has logged out", "id:"+userID.String(), "INFO")
}
