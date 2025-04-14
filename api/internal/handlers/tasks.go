package handlers

import (
	loggergrpc "api/internal/loggerGRPC"
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
		response.WriteAPIResponse(w, http.StatusBadRequest, false, "incomplete headers", nil)
		return
	}

	studentID, err := uuid.Parse(userID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, "invalid student ID", nil)
		loggergrpc.LC.LogInfo("task", "invalid student UUID", map[string]string{"ID": userID})
		return
	}

	student, err := p.User.FindUser(studentID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusNotFound, false, "student not found", nil)
		loggergrpc.LC.LogInfo("task", "student not found", map[string]string{"ID": studentID.String()})
		return
	}

	taskData, err := io.ReadAll(r.Body)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, "failed to read body", nil)
		loggergrpc.LC.LogError("task", "read body failed", map[string]string{"details": err.Error()})
		return
	}

	teacherID := middleware.GetContext(r.Context())

	teacher, err := p.User.FindUser(teacherID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, "teacher not found", nil)
		loggergrpc.LC.LogError("task", "teacher not found", map[string]string{"ID": teacherID.String(), "details": err.Error()})
		return
	}

	taskID := p.Tasks.CreateTask(teacherID, studentID, taskName, teacher.Fio, student.Fio)
	p.Tasks.LinkFileTask(taskID, fileName, taskData)

	response.WriteAPIResponse(w, http.StatusCreated, true, "task was created", "id: "+taskID.String())
	loggergrpc.LC.LogInfo("task", "task created", map[string]string{"task_id": taskID.String(), "student_id": studentID.String(), "teacher_id": teacherID.String()})
}

func (p *TaskHandler) DownloadTask(w http.ResponseWriter, r *http.Request) {
	taskIDStr := r.URL.Query().Get("taskID")
	if taskIDStr == "" {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, "incomplete query", nil)
		return
	}

	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, "invalid task ID", nil)
		loggergrpc.LC.LogInfo("task", "invalid task UUID", map[string]string{"ID": taskIDStr})
		return
	}

	fileName, fileData, err := p.Tasks.GetTask(taskID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusNotFound, false, "file not found", nil)
		loggergrpc.LC.LogInfo("task", "file not found", map[string]string{"task_id": taskID.String()})
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
		response.WriteAPIResponse(w, http.StatusBadRequest, false, "incomplete query", nil)
		return
	}

	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, "invalid task ID", nil)
		loggergrpc.LC.LogInfo("task", "invalid task UUID", map[string]string{"ID": taskIDStr})
		return
	}

	fileName, fileData, err := p.Tasks.GetSolution(taskID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusNotFound, false, "file not found", nil)
		loggergrpc.LC.LogInfo("task", "solution file not found", map[string]string{"task_id": taskID.String()})
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
		response.WriteAPIResponse(w, http.StatusBadRequest, false, "incomplete headers", nil)
		return
	}

	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, "invalid task ID", nil)
		loggergrpc.LC.LogInfo("task", "invalid task UUID", map[string]string{"ID": taskIDStr})
		return
	}

	taskData, err := io.ReadAll(r.Body)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, "failed to read body", nil)
		loggergrpc.LC.LogError("task", "read body failed", map[string]string{"details": err.Error()})
		return
	}

	p.Tasks.LinkFileSolution(taskID, fileName, taskData)
	p.Tasks.Solve(taskID)

	response.WriteAPIResponse(w, http.StatusCreated, true, "task was updated", "id: "+taskID.String())
	loggergrpc.LC.LogInfo("task", "solution added", map[string]string{"task_id": taskID.String()})
}

func (p *TaskHandler) AddGrade(w http.ResponseWriter, r *http.Request) {
	taskIDStr := r.URL.Query().Get("taskID")
	grade := r.URL.Query().Get("grade")
	if taskIDStr == "" || grade == "" {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, "incomplete query", nil)
		return
	}

	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, "invalid task ID", nil)
		loggergrpc.LC.LogInfo("task", "invalid task UUID", map[string]string{"ID": taskIDStr})
		return
	}

	numGrade, err := strconv.Atoi(grade)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, "invalid grade", err.Error())
		loggergrpc.LC.LogInfo("task", "invalid grade format", map[string]string{"grade": grade})
		return
	}

	studentID, err := p.Tasks.Grade(taskID, uint8(numGrade))
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, "could not grade task", nil)
		loggergrpc.LC.LogError("task", "failed to set grade", map[string]string{
			"task_id": taskID.String(), "grade": grade, "details": err.Error(),
		})
		return
	}

	gradeTotal, err := p.Tasks.AvgGrade(studentID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, "could not calculate average grade", nil)
		loggergrpc.LC.LogError("task", "failed to get avg grade", map[string]string{
			"student_id": studentID.String(), "details": err.Error(),
		})
		return
	}

	err = p.User.EditGrade(studentID, gradeTotal)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, "could not update student grade", nil)
		loggergrpc.LC.LogError("task", "failed to update user grade", map[string]string{
			"student_id": studentID.String(), "details": err.Error(),
		})
		return
	}

	response.WriteAPIResponse(w, http.StatusCreated, true, "task was updated", "id: "+taskID.String())
	loggergrpc.LC.LogInfo("task", "grade added", map[string]string{
		"task_id":     taskID.String(),
		"student_id":  studentID.String(),
		"grade":       strconv.Itoa(numGrade),
		"grade_total": strconv.Itoa(int(gradeTotal)),
	})
}

func (p *TaskHandler) OutAllTasks(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetContext(r.Context())

	response.WriteAPIResponse(w, http.StatusOK, true, "", p.Tasks.AllTasks(userID))
}
