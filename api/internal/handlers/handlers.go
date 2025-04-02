package handlers

import (
	"api/internal/encryption"
	"api/internal/response"
	"encoding/base64"
	"html/template"
	"net/http"
)

func EncryptionKey(w http.ResponseWriter, r *http.Request) {
	key, err := encryption.GetEncryptionKey()
	if err != nil {
		response.APIRespond(w, http.StatusInternalServerError, err.Error(), "", "ERROR")
		return
	}
	w.Write([]byte(base64.StdEncoding.EncodeToString(key)))
}

func serveHTML(w http.ResponseWriter, _ *http.Request, filename string) {
	tmpl, err := template.ParseFiles("assets/html/" + filename)
	if err != nil {
		response.APIRespond(w, http.StatusInternalServerError, err.Error(), "", "ERROR")
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
	tmpl, err := template.ParseFiles("assets/html/chat.html")
	if err != nil {
		response.APIRespond(w, http.StatusInternalServerError, err.Error(), "", "ERROR")
		return
	}
	tmpl.Execute(w, nil)
}
