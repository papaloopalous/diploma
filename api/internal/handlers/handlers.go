package handlers

import (
	"api/internal/encryption"
	loggergrpc "api/internal/loggerGRPC"
	"api/internal/response"
	"encoding/base64"
	"html/template"
	"net/http"
)

func EncryptionKey(w http.ResponseWriter, r *http.Request) {
	key, err := encryption.GetEncryptionKey()
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, "encryption error", nil)
		loggergrpc.LC.LogError("encryption", "failed to get an encryption key", map[string]string{"details": err.Error()})
		return
	}
	w.Write([]byte(base64.StdEncoding.EncodeToString(key)))
}

func serveHTML(w http.ResponseWriter, _ *http.Request, filename string) {
	tmpl, err := template.ParseFiles("assets/html/" + filename)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, "failed to output the page", nil)
		loggergrpc.LC.LogError("encryption", "failed to parse the html", map[string]string{"details": err.Error()})
		return
	}
	tmpl.Execute(w, nil)
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
