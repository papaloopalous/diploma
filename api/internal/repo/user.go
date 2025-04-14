package repo

import (
	"errors"
	"sort"

	"github.com/google/uuid"
)

type UserData struct {
	fio       []string
	username  []string
	pass      []string
	role      []string
	age       []uint8
	specialty []string
	id        []uuid.UUID
	price     []int
	teachers  [][]uuid.UUID
	students  [][]uuid.UUID
	rating    []float32
	requests  [][]uuid.UUID
}

type UsersList struct {
	ID        uuid.UUID `json:"id"`
	Fio       string    `json:"fio"`
	Age       uint8     `json:"age"`
	Specialty string    `json:"specialty,omitempty"`
	Price     int       `json:"price,omitempty"`
	Rating    float32   `json:"rating"`
}

type UserRepo interface {
	FindUser(userID uuid.UUID) (user UsersList, err error)
	CheckPass(username string, pass string) (userID uuid.UUID, role string, err error)
	CreateAccount(username string, pass string, role string) (userID uuid.UUID, err error)
	OutAscendingBySpecialty(orderField string, specialty string, userID uuid.UUID) (users []UsersList)
	OutDescendingBySpecialty(orderField string, specialty string, userID uuid.UUID) (users []UsersList)
	HasThatTeacher(studentID uuid.UUID, teacherID uuid.UUID) bool
	AddRating(userID uuid.UUID, rating uint8) error
	StudentsByTeacher(teacherID uuid.UUID) (users []UsersList, err error)
	EditGrade(studentID uuid.UUID, grade float32) error
	FillProfile(userID uuid.UUID, userData UsersList) error
	TeachersByStudent(studentID uuid.UUID) (teachers []UsersList, err error)
	AddRequest(studentID uuid.UUID, teacherID uuid.UUID) error
	ShowRequests(userID uuid.UUID) (users []UsersList, err error)
	Accept(teacherID uuid.UUID, studentID uuid.UUID) error
	Deny(teacherID uuid.UUID, studentID uuid.UUID) error
}

var _ UserRepo = &UserData{}

func NewUserRepo() *UserData {
	return &UserData{
		fio:       make([]string, 0),
		username:  make([]string, 0),
		pass:      make([]string, 0),
		role:      make([]string, 0),
		age:       make([]uint8, 0),
		specialty: make([]string, 0),
		id:        make([]uuid.UUID, 0),
	}
}

func (p *UserData) CreateAccount(username string, pass string, role string) (userID uuid.UUID, err error) {
	for _, val := range p.username {
		if val == username {
			return userID, errors.New("username has been taken")
		}
	}

	userID = uuid.New()

	p.username = append(p.username, username)
	p.role = append(p.role, role)
	p.pass = append(p.pass, pass)
	p.id = append(p.id, userID)
	p.age = append(p.age, 0)
	p.fio = append(p.fio, "")
	p.price = append(p.price, 0)
	p.rating = append(p.rating, 0)
	p.specialty = append(p.specialty, "")
	p.teachers = append(p.teachers, []uuid.UUID{})
	p.students = append(p.students, []uuid.UUID{})
	p.requests = append(p.requests, []uuid.UUID{})

	return userID, nil
}

func (p *UserData) CheckPass(username string, pass string) (userID uuid.UUID, role string, err error) {
	for i, val := range p.username {
		if val == username && pass == p.pass[i] {
			userID = p.id[i]
			role := p.role[i]
			return userID, role, nil
		}
	}

	return userID, role, errors.New("username or password is incorrect")
}

func (p *UserData) FindUser(userID uuid.UUID) (user UsersList, err error) {
	for i, val := range p.id {
		if val == userID {
			user = UsersList{
				ID:        p.id[i],
				Fio:       p.fio[i],
				Age:       p.age[i],
				Specialty: p.specialty[i],
				Price:     p.price[i],
				Rating:    p.rating[i],
			}
			return user, nil
		}
	}

	return user, errors.New("user could not be found")
}

func (p *UserData) OutBySpecialty(orderField, specialty string, studentID uuid.UUID, ascending bool) (users []UsersList) {
	teachers := make(map[uuid.UUID]struct{})
	requests := make(map[uuid.UUID]struct{})

	for i, id := range p.id {
		if id == studentID {
			for _, t := range p.teachers[i] {
				teachers[t] = struct{}{}
			}
			for _, r := range p.requests[i] {
				requests[r] = struct{}{}
			}
			break
		}
	}

	for i := range p.fio {
		if p.role[i] != "teacher" {
			continue
		}
		if specialty != "" && p.specialty[i] != specialty {
			continue
		}
		if _, isTeacher := teachers[p.id[i]]; isTeacher {
			continue
		}
		if _, hasRequest := requests[p.id[i]]; hasRequest {
			continue
		}

		users = append(users, UsersList{
			ID:        p.id[i],
			Fio:       p.fio[i],
			Age:       p.age[i],
			Specialty: p.specialty[i],
			Price:     p.price[i],
			Rating:    p.rating[i],
		})
	}

	switch orderField {
	case "price":
		sort.Slice(users, func(i, j int) bool {
			if ascending {
				return users[i].Price < users[j].Price
			}
			return users[i].Price > users[j].Price
		})
	case "rating":
		sort.Slice(users, func(i, j int) bool {
			if ascending {
				return users[i].Rating < users[j].Rating
			}
			return users[i].Rating > users[j].Rating
		})
	}

	return users
}

func (p *UserData) OutAscendingBySpecialty(orderField, specialty string, studentID uuid.UUID) []UsersList {
	return p.OutBySpecialty(orderField, specialty, studentID, true)
}

func (p *UserData) OutDescendingBySpecialty(orderField, specialty string, studentID uuid.UUID) []UsersList {
	return p.OutBySpecialty(orderField, specialty, studentID, false)
}

func (p *UserData) HasThatTeacher(studentID uuid.UUID, teacherID uuid.UUID) bool {
	for i, val := range p.id {
		if val == studentID {
			for _, val1 := range p.teachers[i] {
				if val1 == teacherID {
					return true
				}
			}
		}
	}

	return false
}

func (p *UserData) AddRating(teacherID uuid.UUID, rating uint8) error {
	for i, val := range p.id {
		if val == teacherID {
			p.rating[i] = (p.rating[i] + float32(rating)) / 2
			return nil
		}
	}

	return errors.New("teacher could no be found")
}

func (p *UserData) StudentsByTeacher(teacherID uuid.UUID) (users []UsersList, err error) {
	var students []uuid.UUID

	found := false
	for i, val := range p.id {
		if val == teacherID {
			students = p.students[i]
			found = true
			break
		}
	}

	if !found {
		return users, errors.New("teacher could not be found")
	}

	for _, val := range students {
		for i, val1 := range p.id {
			if val1 == val {
				users = append(users, UsersList{
					ID:     p.id[i],
					Fio:    p.fio[i],
					Age:    p.age[i],
					Rating: p.rating[i],
				})
			}
		}
	}

	return users, nil
}

func (p *UserData) EditGrade(studentID uuid.UUID, grade float32) error {
	for i, val := range p.id {
		if val == studentID && p.role[i] == "student" {
			p.rating[i] = grade
			return nil
		}
	}

	return errors.New("student could not be found")
}

func (p *UserData) AddRequest(studentID uuid.UUID, teacherID uuid.UUID) error {
	found := false
	for i, val := range p.id {
		if val == teacherID && p.role[i] == "teacher" {
			p.requests[i] = append(p.requests[i], studentID)
			found = true
			break
		}
	}

	if !found {
		return errors.New("teacher could not be found")
	}

	for i, val := range p.id {
		if val == studentID && p.role[i] == "student" {
			p.requests[i] = append(p.requests[i], teacherID)
			return nil
		}
	}

	return errors.New("student could not be found")
}

func (p *UserData) Accept(teacherID uuid.UUID, studentID uuid.UUID) error {
	var requests []uuid.UUID
	var index int

	found := false
	for i, val := range p.id {
		if val == teacherID {
			requests = p.requests[i]
			index = i
			found = true
			break
		}
	}

	if !found {
		return errors.New("teacher could not be found")
	}

	found = false
	for i, val := range requests {
		if val == studentID {
			p.students[index] = append(p.students[index], val)
			requests = append(requests[:i], requests[i+1:]...)
			p.requests[index] = requests
			found = true
			break
		}
	}

	if !found {
		return errors.New("student request could not be found")
	}

	found = false
	for i, val := range p.id {
		if val == studentID {
			requests = p.requests[i]
			index = i
			found = true
			break
		}
	}

	if !found {
		return errors.New("student could not be found")
	}

	for i, val := range requests {
		if val == teacherID {
			p.teachers[index] = append(p.teachers[index], val)
			requests = append(requests[:i], requests[i+1:]...)
			p.requests[index] = requests
			return nil
		}
	}

	return errors.New("teacher request could not be found")

}

func (p *UserData) Deny(teacherID uuid.UUID, studentID uuid.UUID) error {
	var requests []uuid.UUID
	var index int

	found := false
	for i, val := range p.id {
		if val == teacherID {
			requests = p.requests[i]
			index = i
			found = true
			break
		}
	}

	if !found {
		return errors.New("teacher could not be found")
	}

	found = false
	for i, val := range requests {
		if val == studentID {
			requests = append(requests[:i], requests[i+1:]...)
			p.requests[index] = requests
			found = true
			break
		}
	}

	if !found {
		return errors.New("student request could not be found")
	}

	found = false
	for i, val := range p.id {
		if val == studentID {
			requests = p.requests[i]
			index = i
			found = true
			break
		}
	}

	if !found {
		return errors.New("student could not be found")
	}

	for i, val := range requests {
		if val == teacherID {
			requests = append(requests[:i], requests[i+1:]...)
			p.requests[index] = requests
			return nil
		}
	}

	return errors.New("teacher request could not be found")

}

func (p *UserData) ShowRequests(userID uuid.UUID) (users []UsersList, err error) {
	var requestList []uuid.UUID

	found := false
	for i, val := range p.id {
		if val == userID {
			requestList = p.requests[i]
			found = true
			break
		}
	}

	if !found {
		return users, errors.New("user could not be found")
	}

	for _, val := range requestList {
		user, _ := p.FindUser(val)

		users = append(users, user)
	}

	return users, nil
}

func (p *UserData) FillProfile(userID uuid.UUID, userData UsersList) error {
	for i, val := range p.id {
		if val == userID {
			p.age[i] = userData.Age
			p.fio[i] = userData.Fio
			p.specialty[i] = userData.Specialty
			p.price[i] = userData.Price
			return nil
		}
	}

	return errors.New("user could not be found")
}

func (p *UserData) TeachersByStudent(studentID uuid.UUID) (teachers []UsersList, err error) {
	var teachersList []uuid.UUID

	found := false
	for i, val := range p.id {
		if val == studentID {
			teachersList = p.teachers[i]
			break
		}
	}

	if !found {
		return teachers, errors.New("student could not be found")
	}

	for i := range teachersList {
		teachers = append(teachers, UsersList{
			ID:        p.id[i],
			Fio:       p.fio[i],
			Age:       p.age[i],
			Specialty: p.specialty[i],
			Price:     p.price[i],
			Rating:    p.rating[i],
		})
	}

	return teachers, nil
}
