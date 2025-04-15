package handlers

import (
	loggergrpc "api/internal/loggerGRPC"
	"api/internal/messages"
	"api/internal/middleware"
	"api/internal/repo"
	"api/internal/response"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

type UserHandler struct {
	User repo.UserRepo
}

func (p *UserHandler) OutAllTeachers(w http.ResponseWriter, r *http.Request) {
	orderBy := r.URL.Query().Get(messages.ReqOrderBy)
	orderField := r.URL.Query().Get(messages.ReqOrderField)
	specialty := r.URL.Query().Get(messages.ReqSpecialty)

	userID := middleware.GetContext(r.Context())

	if orderBy == "desc" {
		response.WriteAPIResponse(w, http.StatusOK, true, "", p.User.OutDescendingBySpecialty(orderField, specialty, userID))
	} else {
		response.WriteAPIResponse(w, http.StatusOK, true, "", p.User.OutAscendingBySpecialty(orderField, specialty, userID))
	}
}

func (p *UserHandler) AddRating(w http.ResponseWriter, r *http.Request) {
	teacherIDStr := r.URL.Query().Get(messages.ReqTeacherID)
	rating := r.URL.Query().Get(messages.ReqRating)

	if teacherIDStr == "" || rating == "" {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ErrNoParams, nil)
		return
	}

	teacherID, err := uuid.Parse(teacherIDStr)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ErrBadTeacherID, nil)
		loggergrpc.LC.LogError(messages.ServiceUsers, messages.ErrParseTeacherID, map[string]string{messages.LogUserID: teacherIDStr})
		return
	}

	studentID := middleware.GetContext(r.Context())
	if !p.User.HasThatTeacher(studentID, teacherID) {
		response.WriteAPIResponse(w, http.StatusNotFound, false, messages.ErrNotTheirStudent, nil)
		return
	}

	numRating, err := strconv.Atoi(rating)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ErrBadRating, err.Error())
		loggergrpc.LC.LogError(messages.ServiceUsers, messages.ErrParseRating, map[string]string{messages.LogRating: rating})
		return
	}

	p.User.AddRating(teacherID, uint8(numRating))

	response.WriteAPIResponse(w, http.StatusOK, true, messages.StatusRated, nil)
	loggergrpc.LC.LogInfo(messages.ServiceUsers, messages.StatusUserRated, map[string]string{
		messages.LogUserID + messages.RoleTeacher: teacherID.String(),
		messages.LogUserID + messages.RoleStudent: studentID.String(),
		messages.LogRating:                        strconv.Itoa(numRating),
	})
}

func (p *UserHandler) OutRequests(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetContext(r.Context())

	requests, err := p.User.ShowRequests(userID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, err.Error(), nil)
		loggergrpc.LC.LogError(messages.ServiceUsers, err.Error(), map[string]string{
			messages.LogUserID:  userID.String(),
			messages.LogDetails: err.Error(),
		})
		return
	}

	response.WriteAPIResponse(w, http.StatusOK, true, "", requests)
}

func (p *UserHandler) OutAllStudents(w http.ResponseWriter, r *http.Request) {
	teacherID := middleware.GetContext(r.Context())

	students, err := p.User.StudentsByTeacher(teacherID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, err.Error(), nil)
		loggergrpc.LC.LogError(messages.ServiceUsers, err.Error(), map[string]string{
			messages.LogUserID + messages.RoleTeacher: teacherID.String(),
			messages.LogDetails:                       err.Error(),
		})
		return
	}

	response.WriteAPIResponse(w, http.StatusOK, true, "", students)
}

func (p *UserHandler) AddRequest(w http.ResponseWriter, r *http.Request) {
	teacherIDStr := r.URL.Query().Get(messages.ReqTeacherID)
	if teacherIDStr == "" {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ErrNoParams, nil)
		return
	}

	teacherID, err := uuid.Parse(teacherIDStr)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ErrBadTeacherID, nil)
		loggergrpc.LC.LogError(messages.ServiceUsers, messages.ErrParseTeacherID, map[string]string{messages.LogUserID: teacherIDStr})
		return
	}

	studentID := middleware.GetContext(r.Context())
	p.User.AddRequest(studentID, teacherID)

	response.WriteAPIResponse(w, http.StatusCreated, true, messages.StatusReqSent, nil)
	loggergrpc.LC.LogInfo(messages.ServiceUsers, messages.StatusUserReqSent, map[string]string{
		messages.LogUserID + messages.RoleStudent: studentID.String(),
		messages.LogUserID + messages.RoleTeacher: teacherID.String(),
	})
}

func (p *UserHandler) ConfirmRequest(w http.ResponseWriter, r *http.Request) {
	studentIDStr := r.URL.Query().Get(messages.ReqStudentID)
	if studentIDStr == "" {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ErrNoParams, nil)
		return
	}

	studentID, err := uuid.Parse(studentIDStr)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ErrBadStudentID, nil)
		loggergrpc.LC.LogError(messages.ServiceUsers, messages.ErrParseStudentID, map[string]string{messages.LogUserID: studentIDStr})
		return
	}

	teacherID := middleware.GetContext(r.Context())
	err = p.User.Accept(teacherID, studentID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, err.Error(), nil)
		loggergrpc.LC.LogError(messages.ServiceUsers, err.Error(), map[string]string{
			messages.LogUserID + messages.RoleTeacher: teacherID.String(),
			messages.LogUserID + messages.RoleStudent: studentID.String(),
			messages.LogDetails:                       err.Error(),
		})
		return
	}

	response.WriteAPIResponse(w, http.StatusCreated, true, messages.StatusReqAccepted, nil)
	loggergrpc.LC.LogInfo(messages.ServiceUsers, messages.StatusUserReqAccepted, map[string]string{
		messages.LogUserID + messages.RoleTeacher: teacherID.String(),
		messages.LogUserID + messages.RoleStudent: studentID.String(),
	})
}

func (p *UserHandler) DenyRequest(w http.ResponseWriter, r *http.Request) {
	studentIDStr := r.URL.Query().Get(messages.ReqStudentID)
	if studentIDStr == "" {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ErrNoParams, nil)
		return
	}

	studentID, err := uuid.Parse(studentIDStr)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ErrBadStudentID, nil)
		loggergrpc.LC.LogError(messages.ServiceUsers, messages.ErrParseStudentID, map[string]string{messages.LogUserID: studentIDStr})
		return
	}

	teacherID := middleware.GetContext(r.Context())
	p.User.Deny(teacherID, studentID)

	response.WriteAPIResponse(w, http.StatusCreated, true, messages.StatusReqDenied, nil)
	loggergrpc.LC.LogInfo(messages.ServiceUsers, messages.StatusUserReqDenied, map[string]string{
		messages.LogUserID + messages.RoleTeacher: teacherID.String(),
		messages.LogUserID + messages.RoleStudent: studentID.String(),
	})
}

func (p *UserHandler) CancelRequest(w http.ResponseWriter, r *http.Request) {
	teacherIDStr := r.URL.Query().Get(messages.ReqTeacherID)
	if teacherIDStr == "" {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ErrNoParams, nil)
		return
	}

	teacherID, err := uuid.Parse(teacherIDStr)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ErrBadTeacherID, nil)
		loggergrpc.LC.LogError(messages.ServiceUsers, messages.ErrParseTeacherID, map[string]string{messages.LogUserID: teacherIDStr})
		return
	}

	studentID := middleware.GetContext(r.Context())
	p.User.Deny(teacherID, studentID)

	response.WriteAPIResponse(w, http.StatusCreated, true, messages.StatusReqCanceled, nil)
	loggergrpc.LC.LogInfo(messages.ServiceUsers, messages.StatusUserReqCanceled, map[string]string{
		messages.LogUserID + messages.RoleTeacher: teacherID.String(),
		messages.LogUserID + messages.RoleStudent: studentID.String(),
	})
}

func (p *UserHandler) FillProfile(w http.ResponseWriter, r *http.Request) {
	var user repo.UsersList

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, messages.ErrBadRequest, err.Error())
		loggergrpc.LC.LogError(messages.ServiceUsers, messages.ErrDecodeRequest, map[string]string{messages.LogDetails: err.Error()})
		return
	}

	userID := middleware.GetContext(r.Context())
	p.User.FillProfile(userID, user)

	response.WriteAPIResponse(w, http.StatusOK, true, messages.StatusUpdated, nil)
	loggergrpc.LC.LogInfo(messages.ServiceUsers, messages.StatusUserUpdated, map[string]string{messages.LogUserID: userID.String()})
}

func (p *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetContext(r.Context())

	user, err := p.User.FindUser(userID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusNotFound, false, err.Error(), nil)
		loggergrpc.LC.LogError(messages.ServiceUsers, err.Error(), map[string]string{messages.LogUserID: userID.String(), messages.LogDetails: err.Error()})
		return
	}

	response.WriteAPIResponse(w, http.StatusOK, true, "", user)
}

func (p *UserHandler) OutMyTeachers(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetContext(r.Context())

	teachers, err := p.User.TeachersByStudent(userID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, err.Error(), nil)
		loggergrpc.LC.LogError(messages.ServiceUsers, err.Error(), map[string]string{
			messages.LogUserID:  userID.String(),
			messages.LogDetails: err.Error(),
		})
		return
	}

	response.WriteAPIResponse(w, http.StatusOK, true, "", teachers)
}
