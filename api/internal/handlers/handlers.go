package handlers

import (
	"api/internal/encryption"
	"api/internal/response"
	"encoding/base64"
	"html/template"
	"net/http"
)

// func AuthHandler(w http.ResponseWriter, r *http.Request) {
// 	var requestData map[string]string
// 	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
// 		response.APIRespond(w, http.StatusBadRequest, "invalid request", "", "ERROR")
// 		return
// 	}

// 	key, err := encryption.GetEncryptionKey()
// 	if err != nil {
// 		response.APIRespond(w, http.StatusInternalServerError, err.Error(), "", "ERROR")
// 		return
// 	}

// 	encryptedUsername := requestData["username"]
// 	encryptedPassword := requestData["password"]

// 	username, err := encryption.DecryptData(encryptedUsername, key)
// 	if err != nil {
// 		response.APIRespond(w, http.StatusInternalServerError, err.Error(), "", "ERROR")
// 		return
// 	}

// 	password, err := encryption.DecryptData(encryptedPassword, key)
// 	if err != nil {
// 		response.APIRespond(w, http.StatusInternalServerError, err.Error(), "", "ERROR")
// 		return
// 	}

// 	fmt.Println("Decrypted Username:", username)
// 	fmt.Println("Decrypted Password:", password)

// 	response.APIRespond(w, http.StatusOK, "user authenticated", "id:UUID", "INFO")
// }

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
