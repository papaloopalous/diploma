package handlers

import (
	"api/internal/middleware"
	"api/internal/repo"
	"api/internal/response"
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

func (p *UserHandler) AddRequest(w http.ResponseWriter, r *http.Request) {}

func (p *UserHandler) ConfirmRequest(w http.ResponseWriter, r *http.Request) {}

func (p *UserHandler) FillProfile(w http.ResponseWriter, r *http.Request) {}
