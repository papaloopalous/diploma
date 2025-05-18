package handlers

import (
	loggergrpc "api/internal/loggerGRPC"
	"api/internal/messages"
	"api/internal/response"
	"html/template"
	"net/http"
)

// serveHTML обрабатывает запрос на отдачу HTML страницы
func serveHTML(w http.ResponseWriter, r *http.Request, filename string) {
	tmpl, err := template.ParseFiles("assets/html/" + filename)
	if err != nil {
		loggergrpc.LC.LogError(messages.ServiceEncryption, messages.LogErrLoadTemplate, map[string]string{
			messages.LogDetails:  err.Error(),
			messages.LogReqPath:  r.URL.Path,
			messages.LogFilename: filename,
		})
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrPageLoad, nil)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		loggergrpc.LC.LogError(messages.ServiceEncryption, messages.LogErrRenderTemplate, map[string]string{
			messages.LogDetails:  err.Error(),
			messages.LogReqPath:  r.URL.Path,
			messages.LogFilename: filename,
		})
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrPageLoad, nil)
		return
	}

	loggergrpc.LC.LogInfo(messages.ServiceEncryption, messages.LogStatusPageServed, map[string]string{
		messages.LogReqPath:  r.URL.Path,
		messages.LogFilename: filename,
	})
}

// OutIndex отдает главную страницу
func OutIndex(w http.ResponseWriter, r *http.Request) {
	serveHTML(w, r, "index.html")
}

// OutRegister отдает страницу регистрации
func OutRegister(w http.ResponseWriter, r *http.Request) {
	serveHTML(w, r, "register.html")
}

// OutLogin отдает страницу входа
func OutLogin(w http.ResponseWriter, r *http.Request) {
	serveHTML(w, r, "login.html")
}

// OutFillProfile отдает страницу заполнения профиля
func OutFillProfile(w http.ResponseWriter, r *http.Request) {
	serveHTML(w, r, "fill-profile.html")
}

// OutMain отдает главную страницу приложения
func OutMain(w http.ResponseWriter, r *http.Request) {
	serveHTML(w, r, "main.html")
}

// OutTask отдает страницу задания
func OutTask(w http.ResponseWriter, r *http.Request) {
	serveHTML(w, r, "task.html")
}

// OutChat отдает страницу чата
func OutChat(w http.ResponseWriter, r *http.Request) {
	serveHTML(w, r, "chat.html")
}
