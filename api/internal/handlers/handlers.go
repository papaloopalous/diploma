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

func OutIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("assets/html/index.html")
	if err != nil {
		response.APIRespond(w, http.StatusInternalServerError, err.Error(), "", "ERROR")
		return
	}
	tmpl.Execute(w, nil)
}

func OutChat(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("assets/html/chat.html")
	if err != nil {
		response.APIRespond(w, http.StatusInternalServerError, err.Error(), "", "ERROR")
		return
	}
	tmpl.Execute(w, nil)
}

func OutTask(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("assets/html/task.html")
	if err != nil {
		response.APIRespond(w, http.StatusInternalServerError, err.Error(), "", "ERROR")
		return
	}
	tmpl.Execute(w, nil)
}
