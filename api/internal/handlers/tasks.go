package handlers

import (
	errlist "api/internal/errList"
	"api/internal/middleware"
	"api/internal/repo"
	"api/internal/response"
	"io"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

type TaskHandler struct {
	User  repo.UserRepo
	Tasks repo.TaskRepo
}

func (p *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("studentID")
	taskName := r.Header.Get("taskName")
	fileName := r.Header.Get("fileName")

	if userID == "" || taskName == "" || fileName == "" {
		response.APIRespond(w, http.StatusBadRequest, errlist.ErrNoHeaders, "", "ERROR")
		return
	}

	studentID := uuid.MustParse(userID)

	_, err := p.User.FindUser(studentID)
	if err != nil {
		response.APIRespond(w, http.StatusNotFound, errlist.ErrUserNotFound, err.Error(), "ERROR")
		return
	}

	taskData, err := io.ReadAll(r.Body)
	if err != nil {
		response.APIRespond(w, http.StatusInternalServerError, errlist.ErrReadingBody, err.Error(), "ERROR")
		return
	}

	teacherID := middleware.GetContext(r.Context())

	teacher, _ := p.User.FindUser(teacherID)
	student, _ := p.User.FindUser(studentID)

	taskID := p.Tasks.CreateTask(teacherID, studentID, taskName, teacher.Fio, student.Fio)
	p.Tasks.LinkFileTask(taskID, fileName, taskData)

	response.APIRespond(w, http.StatusCreated, "task was created", "id: "+taskID.String(), "INFO")
}

func (p *TaskHandler) DownloadTask(w http.ResponseWriter, r *http.Request) {
	taskIDStr := r.URL.Query().Get("taskID")
	if taskIDStr == "" {
		response.APIRespond(w, http.StatusBadRequest, "incomplete query", "", "ERROR")
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

func (p *TaskHandler) DownloadSolution(w http.ResponseWriter, r *http.Request) {
	taskIDStr := r.URL.Query().Get("taskID")
	if taskIDStr == "" {
		response.APIRespond(w, http.StatusBadRequest, "incomplete query", "", "ERROR")
		return
	}

	taskID := uuid.MustParse(taskIDStr)

	fileName, fileData, _, err := p.Tasks.GetSolution(taskID)
	if err != nil {
		response.APIRespond(w, http.StatusNotFound, "file not found", "", "ERROR")
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	w.Write(fileData)
}

func (p *TaskHandler) AddSolution(w http.ResponseWriter, r *http.Request) {
	taskIDStr := r.Header.Get("taskID")
	fileName := r.Header.Get("fileName")

	if taskIDStr == "" || fileName == "" {
		response.APIRespond(w, http.StatusBadRequest, errlist.ErrNoHeaders, "", "ERROR")
		return
	}

	taskID := uuid.MustParse(taskIDStr)

	taskData, err := io.ReadAll(r.Body)
	if err != nil {
		response.APIRespond(w, http.StatusInternalServerError, errlist.ErrReadingBody, err.Error(), "ERROR")
		return
	}

	p.Tasks.LinkFileSolution(taskID, fileName, taskData)
	p.Tasks.Solve(taskID)

	response.APIRespond(w, http.StatusCreated, "task was updated", "id: "+taskID.String(), "INFO")
}

func (p *TaskHandler) AddGrade(w http.ResponseWriter, r *http.Request) {
	taskIDStr := r.URL.Query().Get("taskID")
	grade := r.URL.Query().Get("grade")
	if taskIDStr == "" || grade == "" {
		response.APIRespond(w, http.StatusBadRequest, "incomplete query", "", "ERROR")
		return
	}

	taskID := uuid.MustParse(taskIDStr)

	numGrade, err := strconv.Atoi(grade)
	if err != nil {
		response.APIRespond(w, http.StatusBadRequest, "invalid grade", err.Error(), "ERROR")
		return
	}

	studentID := p.Tasks.Grade(taskID, uint8(numGrade))
	gradeTotal := p.Tasks.AvgGrade(studentID)
	p.User.EditGrade(studentID, gradeTotal)

	response.APIRespond(w, http.StatusCreated, "task was updated", "id: "+taskID.String(), "INFO")
}

func (p *TaskHandler) OutAllTasks(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetContext(r.Context())

	response.RespondWithJSON(w, http.StatusOK, p.Tasks.AllTasks(userID))
}
