package handlers

import (
	"api/internal/encryption"
	errlist "api/internal/errList"
	"api/internal/repo"
	"api/internal/response"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AuthHandler struct {
	User    repo.UserRepo
	Token   repo.TokenRepo
	Session repo.SessionRepo
}

type MyClaims struct {
	SessionID uuid.UUID `json:"sessionID"`
	jwt.RegisteredClaims
}

func (p *AuthHandler) generateJWT(sessionID uuid.UUID) (string, error) {
	claims := MyClaims{
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(p.Token.GetData())
}

func (p *AuthHandler) parseJWT(tokenString string) (*MyClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		return p.Token.GetData(), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*MyClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New(errlist.ErrInvalidToken)
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
		response.APIRespond(w, http.StatusUnauthorized, "failed to authenticate", "id:"+userID.String(), "INFO")
		return
	}

	sessionID := uuid.New()

	token, err := p.generateJWT(sessionID)
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

	token, err := p.generateJWT(sessionID)
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

	token, err := p.parseJWT(cookie.Value)
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
