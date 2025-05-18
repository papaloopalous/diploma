package handlers

import (
	loggergrpc "api/internal/loggerGRPC"
	"api/internal/messages"
	"api/internal/middleware"
	"api/internal/repo"
	"api/internal/response"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

// UserHandler обрабатывает запросы для работы с пользователями
type UserHandler struct {
	User repo.UserRepo // Репозиторий пользователей
}

// OutAllTeachers возвращает список всех преподавателей
func (p *UserHandler) OutAllTeachers(w http.ResponseWriter, r *http.Request) {
	orderBy := r.URL.Query().Get(messages.ReqOrderBy)
	orderField := r.URL.Query().Get(messages.ReqOrderField)
	specialty := r.URL.Query().Get(messages.ReqSpecialty)

	userID := middleware.GetContext(r.Context())

	loggergrpc.LC.LogInfo(messages.ServiceUsers, messages.LogStatusTeacherListRequested, map[string]string{
		messages.LogUserID:  userID.String(),
		messages.LogDetails: fmt.Sprintf("params: order=%s, field=%s, specialty=%s", orderBy, orderField, specialty),
	})

	var (
		users []repo.UsersList
		err   error
	)

	if orderBy == "desc" {
		users, err = p.User.OutDescendingBySpecialty(orderField, specialty, userID)
	} else {
		users, err = p.User.OutAscendingBySpecialty(orderField, specialty, userID)
	}

	if err != nil {
		loggergrpc.LC.LogError(messages.ServiceUsers, messages.LogErrTeacherList, map[string]string{
			messages.LogUserID:  userID.String(),
			messages.LogDetails: err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrTeacherList, nil)
		return
	}

	response.WriteAPIResponse(w, http.StatusOK, true, messages.StatusSuccess, users)

	loggergrpc.LC.LogInfo(messages.ServiceUsers, messages.LogStatusTeacherList, map[string]string{
		messages.LogUserID:  userID.String(),
		messages.LogDetails: fmt.Sprintf("found %d teachers", len(users)),
	})
}

// AddRating добавляет оценку преподавателю
func (p *UserHandler) AddRating(w http.ResponseWriter, r *http.Request) {
	teacherIDStr := r.URL.Query().Get(messages.ReqTeacherID)
	rating := r.URL.Query().Get(messages.ReqRating)

	if teacherIDStr == "" || rating == "" {
		loggergrpc.LC.LogError(messages.ServiceUsers, messages.LogErrMissingParams, map[string]string{
			messages.LogDetails: "missing teacherId or rating",
		})
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrNoParams, nil)
		return
	}

	teacherID, err := uuid.Parse(teacherIDStr)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrBadTeacherID, nil)
		loggergrpc.LC.LogError(messages.ServiceUsers, messages.LogErrParseTeacherID, map[string]string{
			messages.LogUserID: teacherIDStr,
		})
		return
	}

	studentID := middleware.GetContext(r.Context())
	flag, err := p.User.HasThatTeacher(studentID, teacherID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrCheckTeacher, nil)
		loggergrpc.LC.LogError(messages.ServiceUsers, messages.LogErrCheckTeacher, map[string]string{
			messages.LogUserID + messages.RoleTeacher: teacherID.String(),
			messages.LogUserID + messages.RoleStudent: studentID.String(),
			messages.LogDetails:                       err.Error(),
		})
		return
	}

	if !flag {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrNotTheirStudent, nil)
		return
	}

	numRating, err := strconv.Atoi(rating)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrBadRating, nil)
		loggergrpc.LC.LogError(messages.ServiceUsers, messages.LogErrParseRating, map[string]string{
			messages.LogRating:  rating,
			messages.LogDetails: err.Error(),
		})
		return
	}

	err = p.User.AddRating(teacherID, float32(numRating))
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrAddRating, nil)
		loggergrpc.LC.LogError(messages.ServiceUsers, messages.LogErrAddRating, map[string]string{
			messages.LogUserID + messages.RoleTeacher: teacherID.String(),
			messages.LogUserID + messages.RoleStudent: studentID.String(),
			messages.LogDetails:                       err.Error(),
		})
		return
	}

	response.WriteAPIResponse(w, http.StatusOK, true, messages.StatusRated, nil)
	loggergrpc.LC.LogInfo(messages.ServiceUsers, messages.LogStatusRatingAdded, map[string]string{
		messages.LogUserID + messages.RoleTeacher: teacherID.String(),
		messages.LogUserID + messages.RoleStudent: studentID.String(),
		messages.LogRating:                        rating,
	})
}

// OutRequests возвращает список запросов на обучение
func (p *UserHandler) OutRequests(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetContext(r.Context())

	requests, err := p.User.ShowRequests(userID)
	if err != nil {
		loggergrpc.LC.LogError(messages.ServiceUsers, messages.LogErrRequestList, map[string]string{
			messages.LogUserID:  userID.String(),
			messages.LogDetails: err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrRequestList, nil)
		return
	}

	response.WriteAPIResponse(w, http.StatusOK, true, "", requests)
	loggergrpc.LC.LogInfo(messages.ServiceUsers, messages.LogStatusRequestList, map[string]string{
		messages.LogUserID:  userID.String(),
		messages.LogDetails: fmt.Sprintf("found %d requests", len(requests)),
	})
}

// OutAllStudents возвращает список всех студентов
func (p *UserHandler) OutAllStudents(w http.ResponseWriter, r *http.Request) {
	teacherID := middleware.GetContext(r.Context())

	students, err := p.User.StudentsByTeacher(teacherID)
	if err != nil {
		loggergrpc.LC.LogError(messages.ServiceUsers, messages.LogErrStudentList, map[string]string{
			messages.LogUserID + messages.RoleTeacher: teacherID.String(),
			messages.LogDetails:                       err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrStudentList, nil)
		return
	}

	response.WriteAPIResponse(w, http.StatusOK, true, messages.StatusSuccess, students)
	loggergrpc.LC.LogInfo(messages.ServiceUsers, messages.LogStatusStudentList, map[string]string{
		messages.LogUserID + messages.RoleTeacher: teacherID.String(),
		messages.LogDetails:                       fmt.Sprintf("found %d students", len(students)),
	})
}

// AddRequest добавляет запрос на обучение
func (p *UserHandler) AddRequest(w http.ResponseWriter, r *http.Request) {
	teacherIDStr := r.URL.Query().Get(messages.ReqTeacherID)
	if teacherIDStr == "" {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrNoParams, nil)
		return
	}

	teacherID, err := uuid.Parse(teacherIDStr)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrBadTeacherID, nil)
		loggergrpc.LC.LogError(messages.ServiceUsers, messages.LogErrParseTeacherID, map[string]string{messages.LogUserID: teacherIDStr})
		return
	}

	studentID := middleware.GetContext(r.Context())
	err = p.User.AddRequest(studentID, teacherID)
	if err != nil {
		loggergrpc.LC.LogError(messages.ServiceUsers, messages.LogErrAddRequest, map[string]string{
			messages.LogUserID + messages.RoleTeacher: teacherID.String(),
			messages.LogUserID + messages.RoleStudent: studentID.String(),
			messages.LogDetails:                       err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrAddRequest, nil)
		return
	}

	response.WriteAPIResponse(w, http.StatusCreated, true, messages.StatusReqSent, nil)
	loggergrpc.LC.LogInfo(messages.ServiceUsers, messages.LogStatusUserReqSent, map[string]string{
		messages.LogUserID + messages.RoleStudent: studentID.String(),
		messages.LogUserID + messages.RoleTeacher: teacherID.String(),
	})
}

// ConfirmRequest подтверждает запрос на обучение
func (p *UserHandler) ConfirmRequest(w http.ResponseWriter, r *http.Request) {
	studentIDStr := r.URL.Query().Get(messages.ReqStudentID)
	if studentIDStr == "" {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrNoParams, nil)
		return
	}

	studentID, err := uuid.Parse(studentIDStr)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrBadStudentID, nil)
		loggergrpc.LC.LogError(messages.ServiceUsers, messages.LogErrParseStudentID, map[string]string{messages.LogUserID: studentIDStr})
		return
	}

	teacherID := middleware.GetContext(r.Context())
	err = p.User.Accept(teacherID, studentID)
	if err != nil {
		loggergrpc.LC.LogError(messages.ServiceUsers, messages.LogErrAcceptRequest, map[string]string{
			messages.LogUserID + messages.RoleTeacher: teacherID.String(),
			messages.LogUserID + messages.RoleStudent: studentID.String(),
			messages.LogDetails:                       err.Error(),
		})
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrAcceptRequest, nil)
		return
	}

	response.WriteAPIResponse(w, http.StatusCreated, true, messages.StatusReqAccepted, nil)
	loggergrpc.LC.LogInfo(messages.ServiceUsers, messages.LogStatusUserReqAccepted, map[string]string{
		messages.LogUserID + messages.RoleTeacher: teacherID.String(),
		messages.LogUserID + messages.RoleStudent: studentID.String(),
	})
}

// DenyRequest отклоняет запрос на обучение
func (p *UserHandler) DenyRequest(w http.ResponseWriter, r *http.Request) {
	studentIDStr := r.URL.Query().Get(messages.ReqStudentID)
	if studentIDStr == "" {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrNoParams, nil)
		return
	}

	studentID, err := uuid.Parse(studentIDStr)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrBadStudentID, nil)
		loggergrpc.LC.LogError(messages.ServiceUsers, messages.LogErrParseStudentID, map[string]string{messages.LogUserID: studentIDStr})
		return
	}

	teacherID := middleware.GetContext(r.Context())
	err = p.User.Deny(teacherID, studentID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrDenyRequest, nil)
		loggergrpc.LC.LogError(messages.ServiceUsers, messages.LogErrDenyRequest, map[string]string{
			messages.LogUserID + messages.RoleTeacher: teacherID.String(),
			messages.LogUserID + messages.RoleStudent: studentID.String(),
			messages.LogDetails:                       err.Error(),
		})
		return
	}

	response.WriteAPIResponse(w, http.StatusCreated, true, messages.StatusReqDenied, nil)
	loggergrpc.LC.LogInfo(messages.ServiceUsers, messages.LogStatusUserReqDenied, map[string]string{
		messages.LogUserID + messages.RoleTeacher: teacherID.String(),
		messages.LogUserID + messages.RoleStudent: studentID.String(),
	})
}

// CancelRequest отменяет запрос на обучение
func (p *UserHandler) CancelRequest(w http.ResponseWriter, r *http.Request) {
	teacherIDStr := r.URL.Query().Get(messages.ReqTeacherID)
	if teacherIDStr == "" {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrNoParams, nil)
		return
	}

	teacherID, err := uuid.Parse(teacherIDStr)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrBadTeacherID, nil)
		loggergrpc.LC.LogError(messages.ServiceUsers, messages.LogErrParseTeacherID, map[string]string{messages.LogUserID: teacherIDStr})
		return
	}

	studentID := middleware.GetContext(r.Context())
	err = p.User.Deny(teacherID, studentID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrCancelRequest, nil)
		loggergrpc.LC.LogError(messages.ServiceUsers, messages.LogErrCancelRequest, map[string]string{
			messages.LogUserID + messages.RoleTeacher: teacherID.String(),
			messages.LogUserID + messages.RoleStudent: studentID.String(),
			messages.LogDetails:                       err.Error(),
		})
		return
	}

	response.WriteAPIResponse(w, http.StatusCreated, true, messages.StatusReqCanceled, nil)
	loggergrpc.LC.LogInfo(messages.ServiceUsers, messages.LogStatusUserReqCanceled, map[string]string{
		messages.LogUserID + messages.RoleTeacher: teacherID.String(),
		messages.LogUserID + messages.RoleStudent: studentID.String(),
	})
}

// FillProfile обновляет профиль пользователя
func (p *UserHandler) FillProfile(w http.ResponseWriter, r *http.Request) {
	var user repo.UsersList

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ClientErrBadRequest, nil)
		loggergrpc.LC.LogError(messages.ServiceUsers, messages.LogErrDecodeRequest, map[string]string{
			messages.LogDetails: err.Error(),
		})
		return
	}

	userID := middleware.GetContext(r.Context())
	err := p.User.FillProfile(userID, user)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrFillProfile, nil)
		loggergrpc.LC.LogError(messages.ServiceUsers, messages.LogErrFillProfile, map[string]string{
			messages.LogUserID:  userID.String(),
			messages.LogDetails: err.Error(),
		})
		return
	}

	response.WriteAPIResponse(w, http.StatusOK, true, messages.StatusUpdated, nil)
	loggergrpc.LC.LogInfo(messages.ServiceUsers, messages.LogStatusUserUpdated, map[string]string{messages.LogUserID: userID.String()})
}

func (p *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetContext(r.Context())

	user, err := p.User.FindUser(userID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusNotFound, false, messages.ClientErrFindUser, nil)
		loggergrpc.LC.LogError(messages.ServiceUsers, messages.LogErrFindUser, map[string]string{
			messages.LogUserID:  userID.String(),
			messages.LogDetails: err.Error(),
		})
		return
	}

	response.WriteAPIResponse(w, http.StatusOK, true, "", user)
}

// OutMyTeachers возвращает список преподавателей, с которыми учится студент
func (p *UserHandler) OutMyTeachers(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetContext(r.Context())

	teachers, err := p.User.TeachersByStudent(userID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrGetTeachers, nil)
		loggergrpc.LC.LogError(messages.ServiceUsers, messages.LogErrGetTeachers, map[string]string{
			messages.LogUserID:  userID.String(),
			messages.LogDetails: err.Error(),
		})
		return
	}

	response.WriteAPIResponse(w, http.StatusOK, true, "", teachers)
}
