package handlers

import (
	"api/internal/encryption"
	loggergrpc "api/internal/loggerGRPC"
	"api/internal/messages"
	"api/internal/response"
	"encoding/base64"
	"html/template"
	"net/http"
)

func EncryptionKey(w http.ResponseWriter, r *http.Request) {
	key, err := encryption.GetEncryptionKey()
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ErrEncryption, nil)
		loggergrpc.LC.LogError(messages.ServiceEncryption, messages.ErrKey, map[string]string{messages.LogDetails: err.Error()})
		return
	}
	_, err = w.Write([]byte(base64.StdEncoding.EncodeToString(key)))
	if err != nil {
		loggergrpc.LC.LogError(messages.ServiceEncryption, messages.ErrWriteKey, map[string]string{messages.LogDetails: err.Error()})
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ErrWriteKey, nil)
		return
	}
}

func serveHTML(w http.ResponseWriter, _ *http.Request, filename string) {
	tmpl, err := template.ParseFiles("assets/html/" + filename)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ErrPageOut, nil)
		loggergrpc.LC.LogError(messages.ServiceEncryption, messages.ErrHTML, map[string]string{messages.LogDetails: err.Error()})
		return
	}
	err = tmpl.Execute(w, nil)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ErrPageOut, nil)
		loggergrpc.LC.LogError(messages.ServiceEncryption, messages.ErrHTML, map[string]string{messages.LogDetails: err.Error()})
		return
	}
}

func OutIndex(w http.ResponseWriter, r *http.Request) {
	serveHTML(w, r, "index.html")
}

func OutRegister(w http.ResponseWriter, r *http.Request) {
	serveHTML(w, r, "register.html")
}

func OutLogin(w http.ResponseWriter, r *http.Request) {
	serveHTML(w, r, "login.html")
}

func OutFillProfile(w http.ResponseWriter, r *http.Request) {
	serveHTML(w, r, "fill-profile.html")
}

func OutMain(w http.ResponseWriter, r *http.Request) {
	serveHTML(w, r, "main.html")
}

func OutTask(w http.ResponseWriter, r *http.Request) {
	serveHTML(w, r, "task.html")
}

func OutChat(w http.ResponseWriter, r *http.Request) {
	serveHTML(w, r, "chat.html")
}
