package repo

import (
	"api/internal/messages"
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
			return userID, errors.New(messages.ErrNameTaken)
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

	return userID, role, errors.New(messages.ErrCred)
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

	return user, errors.New(messages.ErrUserNotFound)
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
		if p.role[i] != messages.RoleTeacher {
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
			if p.rating[i] == 0 {
				p.rating[i] = float32(rating)
			} else {
				p.rating[i] = (p.rating[i] + float32(rating)) / 2
			}
			return nil
		}
	}

	return errors.New(messages.ErrTeacherNotFound)
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
		return users, errors.New(messages.ErrTeacherNotFound)
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
		if val == studentID && p.role[i] == messages.RoleStudent {
			p.rating[i] = grade
			return nil
		}
	}

	return errors.New(messages.ErrStudentNotFound)
}

func (p *UserData) AddRequest(studentID uuid.UUID, teacherID uuid.UUID) error {
	found := false
	for i, val := range p.id {
		if val == teacherID && p.role[i] == messages.RoleTeacher {
			p.requests[i] = append(p.requests[i], studentID)
			found = true
			break
		}
	}

	if !found {
		return errors.New(messages.ErrTeacherNotFound)
	}

	for i, val := range p.id {
		if val == studentID && p.role[i] == messages.RoleStudent {
			p.requests[i] = append(p.requests[i], teacherID)
			return nil
		}
	}

	return errors.New(messages.ErrStudentNotFound)
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
		return errors.New(messages.ErrTeacherNotFound)
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
		return errors.New(messages.ErrStudentNotFound)
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
		return errors.New(messages.ErrStudentNotFound)
	}

	for i, val := range requests {
		if val == teacherID {
			p.teachers[index] = append(p.teachers[index], val)
			requests = append(requests[:i], requests[i+1:]...)
			p.requests[index] = requests
			return nil
		}
	}

	return errors.New(messages.ErrTeacherReq)

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
		return errors.New(messages.ErrTeacherNotFound)
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
		return errors.New(messages.ErrStudentNotFound)
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
		return errors.New(messages.ErrStudentNotFound)
	}

	for i, val := range requests {
		if val == teacherID {
			requests = append(requests[:i], requests[i+1:]...)
			p.requests[index] = requests
			return nil
		}
	}

	return errors.New(messages.ErrTeacherReq)

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
		return users, errors.New(messages.ErrUserNotFound)
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

	return errors.New(messages.ErrUserNotFound)
}

func (p *UserData) TeachersByStudent(studentID uuid.UUID) (teachers []UsersList, err error) {
	var teachersList []uuid.UUID

	found := false
	for i, val := range p.id {
		if val == studentID {
			teachersList = p.teachers[i]
			found = true
			break
		}
	}

	if !found {
		return teachers, errors.New(messages.ErrStudentNotFound)
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
