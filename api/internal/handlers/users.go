package handlers

import (
	loggergrpc "api/internal/loggerGRPC"
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
	orderBy := r.URL.Query().Get("orderBy")
	orderField := r.URL.Query().Get("orderField")
	specialty := r.URL.Query().Get("specialty")

	userID := middleware.GetContext(r.Context())

	if orderBy == "desc" {
		response.WriteAPIResponse(w, http.StatusOK, true, "", p.User.OutDescendingBySpecialty(orderField, specialty, userID))
	} else {
		response.WriteAPIResponse(w, http.StatusOK, true, "", p.User.OutAscendingBySpecialty(orderField, specialty, userID))
	}
}

func (p *UserHandler) AddRating(w http.ResponseWriter, r *http.Request) {
	teacherIDStr := r.URL.Query().Get("teacherID")
	rating := r.URL.Query().Get("rating")

	if teacherIDStr == "" || rating == "" {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, "incomplete query", nil)
		return
	}

	teacherID, err := uuid.Parse(teacherIDStr)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, "invalid teacher ID", nil)
		loggergrpc.LC.LogInfo("user", "invalid teacher UUID", map[string]string{"ID": teacherIDStr})
		return
	}

	studentID := middleware.GetContext(r.Context())
	if !p.User.HasThatTeacher(studentID, teacherID) {
		response.WriteAPIResponse(w, http.StatusNotFound, false, "you are not their student", nil)
		return
	}

	numRating, err := strconv.Atoi(rating)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, "invalid rating", err.Error())
		loggergrpc.LC.LogInfo("user", "invalid rating format", map[string]string{"rating": rating})
		return
	}

	p.User.AddRating(teacherID, uint8(numRating))

	response.WriteAPIResponse(w, http.StatusOK, true, "rating was edited", "")
	loggergrpc.LC.LogInfo("user", "rating added", map[string]string{
		"teacher_id": teacherID.String(),
		"student_id": studentID.String(),
		"rating":     strconv.Itoa(numRating),
	})
}

func (p *UserHandler) OutRequests(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetContext(r.Context())

	requests, err := p.User.ShowRequests(userID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, "failed to get requests", nil)
		loggergrpc.LC.LogError("user", "failed to get requests", map[string]string{
			"user_id": userID.String(),
			"details": err.Error(),
		})
		return
	}

	response.WriteAPIResponse(w, http.StatusOK, true, "", requests)
}

func (p *UserHandler) OutAllStudents(w http.ResponseWriter, r *http.Request) {
	teacherID := middleware.GetContext(r.Context())

	students, err := p.User.StudentsByTeacher(teacherID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, "failed to get students", nil)
		loggergrpc.LC.LogError("user", "failed to get students", map[string]string{
			"teacher_id": teacherID.String(),
			"details":    err.Error(),
		})
		return
	}

	response.WriteAPIResponse(w, http.StatusOK, true, "", students)
}

func (p *UserHandler) AddRequest(w http.ResponseWriter, r *http.Request) {
	teacherIDStr := r.URL.Query().Get("teacherID")
	if teacherIDStr == "" {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, "incomplete query", nil)
		return
	}

	teacherID, err := uuid.Parse(teacherIDStr)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, "invalid teacher ID", nil)
		loggergrpc.LC.LogInfo("user", "invalid teacher UUID", map[string]string{"ID": teacherIDStr})
		return
	}

	studentID := middleware.GetContext(r.Context())
	p.User.AddRequest(studentID, teacherID)

	response.WriteAPIResponse(w, http.StatusCreated, true, "request added", "from "+studentID.String()+" to "+teacherID.String())
	loggergrpc.LC.LogInfo("user", "request added", map[string]string{
		"student_id": studentID.String(),
		"teacher_id": teacherID.String(),
	})
}

func (p *UserHandler) ConfirmRequest(w http.ResponseWriter, r *http.Request) {
	studentIDStr := r.URL.Query().Get("studentID")
	if studentIDStr == "" {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, "incomplete query", nil)
		return
	}

	studentID, err := uuid.Parse(studentIDStr)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, "invalid student ID", nil)
		loggergrpc.LC.LogInfo("user", "invalid student UUID", map[string]string{"ID": studentIDStr})
		return
	}

	teacherID := middleware.GetContext(r.Context())
	err = p.User.Accept(teacherID, studentID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, "could not confirm request", nil)
		loggergrpc.LC.LogError("user", "failed to confirm request", map[string]string{
			"teacher_id": teacherID.String(),
			"student_id": studentID.String(),
			"details":    err.Error(),
		})
		return
	}

	response.WriteAPIResponse(w, http.StatusCreated, true, "request accepted", "from "+studentID.String()+" to "+teacherID.String())
	loggergrpc.LC.LogInfo("user", "request confirmed", map[string]string{
		"teacher_id": teacherID.String(),
		"student_id": studentID.String(),
	})
}

func (p *UserHandler) DenyRequest(w http.ResponseWriter, r *http.Request) {
	studentIDStr := r.URL.Query().Get("studentID")
	if studentIDStr == "" {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, "incomplete query", nil)
		return
	}

	studentID, err := uuid.Parse(studentIDStr)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, "invalid student ID", nil)
		loggergrpc.LC.LogInfo("user", "invalid student UUID", map[string]string{"ID": studentIDStr})
		return
	}

	teacherID := middleware.GetContext(r.Context())
	p.User.Deny(teacherID, studentID)

	response.WriteAPIResponse(w, http.StatusCreated, true, "request denied", "from "+studentID.String()+" to "+teacherID.String())
	loggergrpc.LC.LogInfo("user", "request denied", map[string]string{
		"teacher_id": teacherID.String(),
		"student_id": studentID.String(),
	})
}

func (p *UserHandler) CancelRequest(w http.ResponseWriter, r *http.Request) {
	teacherIDStr := r.URL.Query().Get("teacherID")
	if teacherIDStr == "" {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, "incomplete query", nil)
		return
	}

	teacherID, err := uuid.Parse(teacherIDStr)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, "invalid teacher ID", nil)
		loggergrpc.LC.LogInfo("user", "invalid teacher UUID", map[string]string{"ID": teacherIDStr})
		return
	}

	studentID := middleware.GetContext(r.Context())
	p.User.Deny(teacherID, studentID)

	response.WriteAPIResponse(w, http.StatusCreated, true, "request canceled", "from "+studentID.String()+" to "+teacherID.String())
	loggergrpc.LC.LogInfo("user", "request canceled", map[string]string{
		"teacher_id": teacherID.String(),
		"student_id": studentID.String(),
	})
}

func (p *UserHandler) FillProfile(w http.ResponseWriter, r *http.Request) {
	var user repo.UsersList

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		response.WriteAPIResponse(w, http.StatusBadRequest, false, "failed to decode user", err.Error())
		loggergrpc.LC.LogError("user", "failed to decode profile", map[string]string{"details": err.Error()})
		return
	}

	userID := middleware.GetContext(r.Context())
	p.User.FillProfile(userID, user)

	response.WriteAPIResponse(w, http.StatusOK, true, "profile updated", "id: "+userID.String())
	loggergrpc.LC.LogInfo("user", "profile updated", map[string]string{"user_id": userID.String()})
}

func (p *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetContext(r.Context())

	user, err := p.User.FindUser(userID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusNotFound, false, "user not found", nil)
		loggergrpc.LC.LogInfo("user", "user not found", map[string]string{"user_id": userID.String(), "details": err.Error()})
		return
	}

	response.WriteAPIResponse(w, http.StatusOK, true, "", user)
}

func (p *UserHandler) OutMyTeachers(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetContext(r.Context())

	teachers, err := p.User.TeachersByStudent(userID)
	if err != nil {
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, "failed to get teachers", nil)
		loggergrpc.LC.LogError("user", "failed to get teachers", map[string]string{
			"user_id": userID.String(),
			"details": err.Error(),
		})
		return
	}

	response.WriteAPIResponse(w, http.StatusOK, true, "", teachers)
}
