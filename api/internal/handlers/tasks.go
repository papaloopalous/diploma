package handlers

import (
	loggergrpc "api/internal/loggerGRPC"
	"api/internal/messages"
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
	userID := r.Header.Get(messages.ReqStudentID)
	taskName := r.Header.Get(messages.ReqTaskName)
	fileName := r.Header.Get(messages.ReqFileName)

	if userID == "" || taskName == "" || fileName == "" {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ErrNoParams, nil)
		return
	}

	studentID, err := uuid.Parse(userID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ErrBadStudentID, nil)
		loggergrpc.LC.LogError(messages.ServiceTasks, messages.ErrParseStudentID, map[string]string{messages.LogUserID: userID})
		return
	}

	student, err := p.User.FindUser(studentID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusNotFound, false, err.Error(), nil)
		loggergrpc.LC.LogError(messages.ServiceTasks, err.Error(), map[string]string{messages.LogUserID: studentID.String()})
		return
	}

	taskData, err := io.ReadAll(r.Body)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ErrBadRequest, nil)
		loggergrpc.LC.LogError(messages.ServiceTasks, messages.ErrDecodeRequest, map[string]string{messages.LogDetails: err.Error()})
		return
	}

	teacherID := middleware.GetContext(r.Context())

	teacher, err := p.User.FindUser(teacherID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, err.Error(), nil)
		loggergrpc.LC.LogError(messages.ServiceTasks, err.Error(), map[string]string{
			messages.LogUserID:  teacherID.String(),
			messages.LogDetails: err.Error(),
		})
		return
	}

	taskID, err := p.Tasks.CreateTask(teacherID, studentID, taskName, teacher.Fio, student.Fio)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, err.Error(), nil)
		loggergrpc.LC.LogError(messages.ServiceTasks, err.Error(), map[string]string{
			messages.LogUserID:  teacherID.String(),
			messages.LogDetails: err.Error(),
		})
		return
	}

	err = p.Tasks.LinkFileTask(taskID, fileName, taskData)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, err.Error(), nil)
		loggergrpc.LC.LogError(messages.ServiceTasks, err.Error(), map[string]string{
			messages.LogUserID:  teacherID.String(),
			messages.LogDetails: err.Error(),
		})
		return
	}

	response.WriteAPIResponse(w, http.StatusCreated, true, messages.StatusTaskCreated, "id: "+taskID.String())
	loggergrpc.LC.LogInfo(messages.ServiceTasks, messages.StatusUserTaskCreated, map[string]string{
		messages.LogTaskID:                        taskID.String(),
		messages.LogUserID + messages.RoleStudent: studentID.String(),
		messages.LogUserID + messages.RoleTeacher: teacherID.String(),
	})
}

func (p *TaskHandler) DownloadTask(w http.ResponseWriter, r *http.Request) {
	taskIDStr := r.URL.Query().Get(messages.ReqTaskID)
	if taskIDStr == "" {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ErrNoParams, nil)
		return
	}

	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ErrBadTaskID, nil)
		loggergrpc.LC.LogError(messages.ServiceTasks, messages.ErrParseTaskID, map[string]string{messages.LogTaskID: taskIDStr})
		return
	}

	fileName, fileData, err := p.Tasks.GetTask(taskID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusNotFound, false, err.Error(), nil)
		loggergrpc.LC.LogError(messages.ServiceTasks, err.Error(), map[string]string{messages.LogTaskID: taskID.String()})
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(fileData)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ErrWriteFile, nil)
		loggergrpc.LC.LogError(messages.ServiceTasks, messages.ErrWriteFile, map[string]string{messages.LogDetails: err.Error()})
		return
	}
}

func (p *TaskHandler) DownloadSolution(w http.ResponseWriter, r *http.Request) {
	taskIDStr := r.URL.Query().Get(messages.ReqTaskID)
	if taskIDStr == "" {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ErrNoParams, nil)
		return
	}

	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ErrBadTaskID, nil)
		loggergrpc.LC.LogError(messages.ServiceTasks, messages.ErrParseTaskID, map[string]string{messages.LogTaskID: taskIDStr})
		return
	}

	fileName, fileData, err := p.Tasks.GetSolution(taskID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusNotFound, false, err.Error(), nil)
		loggergrpc.LC.LogError(messages.ServiceTasks, err.Error(), map[string]string{messages.LogTaskID: taskID.String()})
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(fileData)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ErrWriteFile, nil)
		loggergrpc.LC.LogError(messages.ServiceTasks, messages.ErrWriteFile, map[string]string{messages.LogDetails: err.Error()})
		return
	}
}

func (p *TaskHandler) AddSolution(w http.ResponseWriter, r *http.Request) {
	taskIDStr := r.Header.Get(messages.ReqTaskID)
	fileName := r.Header.Get(messages.ReqFileName)

	if taskIDStr == "" || fileName == "" {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ErrNoParams, nil)
		return
	}

	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ErrBadTaskID, nil)
		loggergrpc.LC.LogError(messages.ServiceTasks, messages.ErrParseTaskID, map[string]string{messages.LogTaskID: taskIDStr})
		return
	}

	taskData, err := io.ReadAll(r.Body)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ErrBadRequest, nil)
		loggergrpc.LC.LogError(messages.ServiceTasks, messages.ErrDecodeRequest, map[string]string{messages.LogDetails: err.Error()})
		return
	}

	err = p.Tasks.LinkFileSolution(taskID, fileName, taskData)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, err.Error(), nil)
		loggergrpc.LC.LogError(messages.ServiceTasks, err.Error(), map[string]string{messages.LogTaskID: taskID.String()})
		return
	}

	err = p.Tasks.Solve(taskID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, err.Error(), nil)
		loggergrpc.LC.LogError(messages.ServiceTasks, err.Error(), map[string]string{messages.LogTaskID: taskID.String()})
		return
	}

	response.WriteAPIResponse(w, http.StatusCreated, true, messages.StatusTaskUpdated, "id: "+taskID.String())
	loggergrpc.LC.LogInfo(messages.ServiceTasks, messages.StatusUserSolution, map[string]string{messages.LogTaskID: taskID.String()})
}

func (p *TaskHandler) AddGrade(w http.ResponseWriter, r *http.Request) {
	taskIDStr := r.URL.Query().Get(messages.ReqTaskID)
	grade := r.URL.Query().Get(messages.ReqGrade)
	if taskIDStr == "" || grade == "" {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ErrNoParams, nil)
		return
	}

	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ErrBadTaskID, nil)
		loggergrpc.LC.LogError(messages.ServiceTasks, messages.ErrParseTaskID, map[string]string{messages.LogTaskID: taskIDStr})
		return
	}

	numGrade, err := strconv.Atoi(grade)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ErrBadGrade, err.Error())
		loggergrpc.LC.LogError(messages.ServiceTasks, messages.ErrParseGrade, map[string]string{messages.LogGrade: grade})
		return
	}

	studentID, err := p.Tasks.Grade(taskID, uint8(numGrade))
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, err.Error(), nil)
		loggergrpc.LC.LogError(messages.ServiceTasks, err.Error(), map[string]string{
			messages.LogTaskID: taskID.String(),
			messages.LogGrade:  grade,
		})
		return
	}

	gradeTotal, err := p.Tasks.AvgGrade(studentID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, err.Error(), nil)
		loggergrpc.LC.LogError(messages.ServiceTasks, err.Error(), map[string]string{
			messages.LogUserID: studentID.String(), messages.LogDetails: err.Error(),
		})
		return
	}

	err = p.User.EditGrade(studentID, gradeTotal)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusNotFound, false, err.Error(), nil)
		loggergrpc.LC.LogError(messages.ServiceTasks, err.Error(), map[string]string{
			messages.LogUserID: studentID.String(), messages.LogDetails: err.Error(),
		})
		return
	}

	response.WriteAPIResponse(w, http.StatusCreated, true, messages.StatusTaskUpdated, "id: "+taskID.String())
	loggergrpc.LC.LogInfo(messages.ServiceTasks, messages.StatusUserGrade, map[string]string{
		messages.LogTaskID: taskID.String(),
		messages.LogUserID: studentID.String(),
		messages.LogGrade:  strconv.Itoa(numGrade),
		messages.LogTotal:  strconv.Itoa(int(gradeTotal)),
	})
}

func (p *TaskHandler) OutAllTasks(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetContext(r.Context())

	response.WriteAPIResponse(w, http.StatusOK, true, "", p.Tasks.AllTasks(userID))
}
