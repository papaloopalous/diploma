package handlers

import (
	errlist "api/internal/errList"
	"api/internal/repo"
	"api/internal/response"
	"io"
	"net/http"

	"github.com/google/uuid"
)

type TaskHandler struct {
	User  repo.UserRepo
	Tasks repo.TaskRepo
}

func (p *TaskHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("studentID")
	taskName := r.Header.Get("taskName")
	fileName := r.Header.Get("fileName")

	if userID == "" || taskName == "" || fileName == "" {
		response.APIRespond(w, http.StatusBadRequest, errlist.ErrNoHeaders, "", "ERROR")
		return
	}

	studentID := uuid.MustParse(userID)

	// _, _, _, _, err := p.User.FindUser(studentID)
	// if err != nil {
	// 	response.APIRespond(w, http.StatusNotFound, errlist.ErrUserNotFound, err.Error(), "ERROR")
	// 	return
	// }

	taskData, err := io.ReadAll(r.Body)
	if err != nil {
		response.APIRespond(w, http.StatusInternalServerError, errlist.ErrReadingBody, err.Error(), "ERROR")
		return
	}

	//teacherID := middleware.GetContext(r.Context())
	teacherID := uuid.MustParse("b65bd4af-797c-4a12-927a-fd807bf95b27")
	taskID := p.Tasks.CreateTask(teacherID, studentID, taskName)
	p.Tasks.LinkFile(taskID, fileName, taskData)

	response.APIRespond(w, http.StatusCreated, "task was created", "id: "+taskID.String(), "INFO")
}

func (p *TaskHandler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	taskIDStr := r.URL.Query().Get("taskID")
	if taskIDStr == "" {
		response.APIRespond(w, http.StatusBadRequest, "missing task id", "", "ERROR")
		return
	}

	taskID := uuid.MustParse(taskIDStr)

	fileName, fileData, _, err := p.Tasks.GetTask(taskID)
	if err != nil {
		response.APIRespond(w, http.StatusNotFound, "file not found", "", "ERROR")
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	w.Write(fileData)
}
