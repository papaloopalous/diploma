package handlers

import (
	loggergrpc "api/internal/loggerGRPC"
	"api/internal/messages"
	"api/internal/middleware"
	"api/internal/repo"
	"api/internal/response"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

// TaskHandler обрабатывает запросы для работы с заданиями
type TaskHandler struct {
	User  repo.UserRepo // Репозиторий пользователей
	Tasks repo.TaskRepo // Репозиторий заданий
}

// CreateTask создает новое задание
func (p *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get(messages.ReqStudentID)
	taskName := r.Header.Get(messages.ReqTaskName)
	fileName := r.Header.Get(messages.ReqFileName)

	if userID == "" || taskName == "" || fileName == "" {
		loggergrpc.LC.LogError(messages.ServiceTasks, messages.LogErrParamsRequest, map[string]string{
			messages.LogDetails: messages.LogErrNoParams,
		})
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrBadRequest, nil)
		return
	}

	studentID, err := uuid.Parse(userID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrBadStudentID, nil)
		loggergrpc.LC.LogError(messages.ServiceTasks, messages.LogErrParseStudentID, map[string]string{messages.LogUserID: userID})
		return
	}

	student, err := p.User.FindUser(studentID)
	if err != nil {
		loggergrpc.LC.LogError(messages.ServiceTasks, messages.LogErrUserNotFound, map[string]string{
			messages.LogUserID:  studentID.String(),
			messages.LogDetails: err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusNotFound, false, messages.ClientErrUserNotFound, nil)
		return
	}

	taskData, err := io.ReadAll(r.Body)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrBadRequest, nil)
		loggergrpc.LC.LogError(messages.ServiceTasks, messages.LogErrDecodeRequest, map[string]string{messages.LogDetails: err.Error()})
		return
	}

	teacherID := middleware.GetContext(r.Context())

	teacher, err := p.User.FindUser(teacherID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrFindTeacher, nil)
		loggergrpc.LC.LogError(messages.ServiceTasks, messages.LogErrFindTeacher, map[string]string{
			messages.LogUserID:  teacherID.String(),
			messages.LogDetails: err.Error(),
		})
		return
	}

	taskID, err := p.Tasks.CreateTask(teacherID, studentID, taskName, teacher.Fio, student.Fio)
	if err != nil {
		loggergrpc.LC.LogError(messages.ServiceTasks, messages.LogErrTaskCreate, map[string]string{
			messages.LogUserID + messages.RoleTeacher: teacherID.String(),
			messages.LogUserID + messages.RoleStudent: studentID.String(),
			messages.LogDetails:                       err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrTaskNotFound, nil)
		return
	}

	err = p.Tasks.LinkFileTask(taskID, fileName, taskData)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrLinkFile, nil)
		loggergrpc.LC.LogError(messages.ServiceTasks, messages.LogErrLinkFile, map[string]string{
			messages.LogUserID:  teacherID.String(),
			messages.LogDetails: err.Error(),
		})
		return
	}

	response.WriteAPIResponse(w, http.StatusCreated, true, messages.StatusTaskCreated, map[string]string{
		messages.LogTaskID: taskID.String(),
	})
	loggergrpc.LC.LogInfo(messages.ServiceTasks, messages.LogStatusTaskCreated, map[string]string{
		messages.LogTaskID:                        taskID.String(),
		messages.LogUserID + messages.RoleTeacher: teacherID.String(),
		messages.LogUserID + messages.RoleStudent: studentID.String(),
	})
}

// DownloadTask скачивает файл задания
func (p *TaskHandler) DownloadTask(w http.ResponseWriter, r *http.Request) {
	taskIDStr := r.URL.Query().Get(messages.ReqTaskID)
	if taskIDStr == "" {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrNoParams, nil)
		return
	}

	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrBadTaskID, nil)
		loggergrpc.LC.LogError(messages.ServiceTasks, messages.LogErrParseTaskID, map[string]string{messages.LogTaskID: taskIDStr})
		return
	}

	fileName, fileData, err := p.Tasks.GetTask(taskID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusNotFound, false, messages.ClientErrGetTask, nil)
		loggergrpc.LC.LogError(messages.ServiceTasks, messages.LogErrGetTask, map[string]string{
			messages.LogTaskID:  taskID.String(),
			messages.LogDetails: err.Error(),
		})
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(fileData)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrWriteFile, nil)
		loggergrpc.LC.LogError(messages.ServiceTasks, messages.LogErrWriteFile, map[string]string{messages.LogDetails: err.Error()})
		return
	}

	loggergrpc.LC.LogInfo(messages.ServiceTasks, messages.LogStatusFileDownload, map[string]string{
		messages.LogTaskID:   taskID.String(),
		messages.LogFilename: fileName,
	})
}

// DownloadSolution скачивает файл решения задания
func (p *TaskHandler) DownloadSolution(w http.ResponseWriter, r *http.Request) {
	taskIDStr := r.URL.Query().Get(messages.ReqTaskID)
	if taskIDStr == "" {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrNoParams, nil)
		return
	}

	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrBadTaskID, nil)
		loggergrpc.LC.LogError(messages.ServiceTasks, messages.LogErrParseTaskID, map[string]string{messages.LogTaskID: taskIDStr})
		return
	}

	fileName, fileData, err := p.Tasks.GetSolution(taskID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusNotFound, false, messages.ClientErrGetSolution, nil)
		loggergrpc.LC.LogError(messages.ServiceTasks, messages.LogErrGetSolution, map[string]string{
			messages.LogTaskID:  taskID.String(),
			messages.LogDetails: err.Error(),
		})
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(fileData)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrWriteFile, nil)
		loggergrpc.LC.LogError(messages.ServiceTasks, messages.LogErrWriteFile, map[string]string{messages.LogDetails: err.Error()})
		return
	}

	loggergrpc.LC.LogInfo(messages.ServiceTasks, messages.LogStatusFileDownload, map[string]string{
		messages.LogTaskID:   taskID.String(),
		messages.LogFilename: fileName,
	})
}

// AddSolution добавляет решение задания
func (p *TaskHandler) AddSolution(w http.ResponseWriter, r *http.Request) {
	taskIDStr := r.Header.Get(messages.ReqTaskID)
	fileName := r.Header.Get(messages.ReqFileName)

	if taskIDStr == "" || fileName == "" {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrNoParams, nil)
		return
	}

	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrBadTaskID, nil)
		loggergrpc.LC.LogError(messages.ServiceTasks, messages.LogErrParseTaskID, map[string]string{messages.LogTaskID: taskIDStr})
		return
	}

	taskData, err := io.ReadAll(r.Body)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrBadRequest, nil)
		loggergrpc.LC.LogError(messages.ServiceTasks, messages.LogErrDecodeRequest, map[string]string{messages.LogDetails: err.Error()})
		return
	}

	err = p.Tasks.LinkFileSolution(taskID, fileName, taskData)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrSaveSolution, nil)
		loggergrpc.LC.LogError(messages.ServiceTasks, messages.LogErrSaveSolution, map[string]string{
			messages.LogTaskID:  taskID.String(),
			messages.LogDetails: err.Error(),
		})
		return
	}

	err = p.Tasks.Solve(taskID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrSaveSolution, nil)
		loggergrpc.LC.LogError(messages.ServiceTasks, messages.LogErrSaveSolution, map[string]string{
			messages.LogTaskID:  taskID.String(),
			messages.LogDetails: err.Error(),
		})
		return
	}

	response.WriteAPIResponse(w, http.StatusCreated, true, messages.StatusTaskUpdated, map[string]string{
		messages.LogTaskID: taskID.String(),
	})
	loggergrpc.LC.LogInfo(messages.ServiceTasks, messages.LogStatusSolutionAdded, map[string]string{
		messages.LogTaskID:   taskID.String(),
		messages.LogFilename: fileName,
	})
}

// AddGrade добавляет оценку к заданию
func (p *TaskHandler) AddGrade(w http.ResponseWriter, r *http.Request) {
	taskIDStr := r.URL.Query().Get(messages.ReqTaskID)
	grade := r.URL.Query().Get(messages.ReqGrade)
	if taskIDStr == "" || grade == "" {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrNoParams, nil)
		return
	}

	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrBadTaskID, nil)
		loggergrpc.LC.LogError(messages.ServiceTasks, messages.LogErrParseTaskID, map[string]string{messages.LogTaskID: taskIDStr})
		return
	}

	numGrade, err := strconv.Atoi(grade)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrBadGrade, err.Error())
		loggergrpc.LC.LogError(messages.ServiceTasks, messages.LogErrParseGrade, map[string]string{messages.LogGrade: grade})
		return
	}

	studentID, err := p.Tasks.Grade(taskID, uint8(numGrade))
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrGradeTask, nil)
		loggergrpc.LC.LogError(messages.ServiceTasks, messages.LogErrGradeTask, map[string]string{
			messages.LogTaskID:  taskID.String(),
			messages.LogGrade:   grade,
			messages.LogDetails: err.Error(),
		})
		return
	}

	gradeTotal, err := p.Tasks.AvgGrade(studentID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrCalcGrade, nil)
		loggergrpc.LC.LogError(messages.ServiceTasks, messages.LogErrCalcGrade, map[string]string{
			messages.LogUserID:  studentID.String(),
			messages.LogDetails: err.Error(),
		})
		return
	}

	err = p.User.EditGrade(studentID, gradeTotal)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusNotFound, false, messages.ClientErrUpdateGrade, nil)
		loggergrpc.LC.LogError(messages.ServiceTasks, messages.LogErrUpdateGrade, map[string]string{
			messages.LogUserID:  studentID.String(),
			messages.LogDetails: err.Error(),
		})
		return
	}

	response.WriteAPIResponse(w, http.StatusOK, true, messages.StatusGradeAdded, map[string]string{
		messages.LogTaskID: taskID.String(),
	})
	loggergrpc.LC.LogInfo(messages.ServiceTasks, messages.LogStatusGradeAdded, map[string]string{
		messages.LogTaskID: taskID.String(),
		messages.LogUserID: studentID.String(),
		messages.LogGrade:  grade,
		messages.LogTotal:  strconv.FormatFloat(float64(gradeTotal), 'f', 2, 64),
	})
}

// OutAllTasks выводит все задания пользователя
func (p *TaskHandler) OutAllTasks(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetContext(r.Context())
	tasks := p.Tasks.AllTasks(userID)

	response.WriteAPIResponse(w, http.StatusOK, true, messages.StatusSuccess, tasks)
	loggergrpc.LC.LogInfo(messages.ServiceTasks, messages.LogStatusTaskList, map[string]string{
		messages.LogUserID:  userID.String(),
		messages.LogDetails: fmt.Sprintf("found %d tasks", len(tasks)),
	})
}
