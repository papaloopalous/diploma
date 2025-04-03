package handlers

import (
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

	if orderBy == "desc" {
		response.RespondWithJSON(w, http.StatusOK, p.User.OutDescendingBySpecialty(orderField, specialty))
	} else {
		response.RespondWithJSON(w, http.StatusOK, p.User.OutAscendingBySpecialty(orderField, specialty))
	}
}

func (p *UserHandler) AddRating(w http.ResponseWriter, r *http.Request) {
	teacherIDStr := r.URL.Query().Get("teacherID")
	rating := r.URL.Query().Get("rating")

	if teacherIDStr == "" {
		response.APIRespond(w, http.StatusBadRequest, "incomplete query", "", "ERROR")
		return
	}

	teacherID := uuid.MustParse(teacherIDStr)

	studentID := middleware.GetContext(r.Context())
	flag := p.User.HasThatTeacher(studentID, teacherID)

	if !flag {
		response.APIRespond(w, http.StatusNotFound, "you are not their student", "", "INFO")
		return
	}

	numRating, err := strconv.Atoi(rating)
	if err != nil {
		response.APIRespond(w, http.StatusBadRequest, "invalid rating", err.Error(), "ERROR")
		return
	}

	p.User.AddRating(teacherID, uint8(numRating))

	response.APIRespond(w, http.StatusOK, "rating was edited", "", "INFO")
}

func (p *UserHandler) OutAllStudents(w http.ResponseWriter, r *http.Request) {
	teacherID := middleware.GetContext(r.Context())

	response.RespondWithJSON(w, http.StatusOK, p.User.StudentsByTeacher(teacherID))
}

func (p *UserHandler) AddRequest(w http.ResponseWriter, r *http.Request) {
	teacherIDStr := r.URL.Query().Get("teacherID")
	if teacherIDStr == "" {
		response.APIRespond(w, http.StatusBadRequest, "incomplete query", "", "ERROR")
		return
	}

	teacherID := uuid.MustParse(teacherIDStr)
	studentID := middleware.GetContext(r.Context())

	p.User.AddRequest(studentID, teacherID)

	response.APIRespond(w, http.StatusCreated, "request added", "from "+studentID.String()+" to "+teacherID.String(), "INFO")
}

func (p *UserHandler) ConfirmRequest(w http.ResponseWriter, r *http.Request) {
	studentIDStr := r.URL.Query().Get("studentID")
	if studentIDStr == "" {
		response.APIRespond(w, http.StatusBadRequest, "incomplete query", "", "ERROR")
		return
	}

	studentID := uuid.MustParse(studentIDStr)
	teacherID := middleware.GetContext(r.Context())

	p.User.Accept(teacherID, studentID)

	response.APIRespond(w, http.StatusCreated, "request accepted", "from "+studentID.String()+" to "+teacherID.String(), "INFO")
}

func (p *UserHandler) DenyRequest(w http.ResponseWriter, r *http.Request) {
	studentIDStr := r.URL.Query().Get("studentID")
	if studentIDStr == "" {
		response.APIRespond(w, http.StatusBadRequest, "incomplete query", "", "ERROR")
		return
	}

	studentID := uuid.MustParse(studentIDStr)
	teacherID := middleware.GetContext(r.Context())

	p.User.Deny(teacherID, studentID)

	response.APIRespond(w, http.StatusCreated, "request denied", "from "+studentID.String()+" to "+teacherID.String(), "INFO")
}

func (p *UserHandler) OutRequests(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetContext(r.Context())

	response.RespondWithJSON(w, http.StatusOK, p.User.ShowRequests(userID))
}

func (p *UserHandler) FillProfile(w http.ResponseWriter, r *http.Request) {
	var user repo.UsersList

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		response.APIRespond(w, http.StatusBadRequest, "failde to decode user", err.Error(), "ERROR")
		return
	}

	userID := middleware.GetContext(r.Context())

	p.User.FillProfile(userID, user)

	response.APIRespond(w, http.StatusOK, "profile updated", "id: "+userID.String(), "INFO")
}

func (p *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetContext(r.Context())

	user, err := p.User.FindUser(userID)
	if err != nil {
		response.APIRespond(w, http.StatusNotFound, err.Error(), "", "ERROR")
		return
	}

	response.RespondWithJSON(w, http.StatusOK, user)
}

func (p *UserHandler) OutMyTeachers(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetContext(r.Context())

	response.RespondWithJSON(w, http.StatusOK, p.User.TeachersByStudent(userID))
}
