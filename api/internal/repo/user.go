package repo

import (
	errlist "api/internal/errList"
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
	OutAscendingBySpecialty(orderField string, specialty string) (users []UsersList)
	OutDescendingBySpecialty(orderField string, specialty string) (users []UsersList)
	HasThatTeacher(studentID uuid.UUID, teacherID uuid.UUID) bool
	AddRating(userID uuid.UUID, rating uint8)
	StudentsByTeacher(teacherID uuid.UUID) (users []UsersList)
	EditGrade(studentID uuid.UUID, grade float32)
	FillProfile(userID uuid.UUID, userData UsersList)
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
			return userID, errors.New(errlist.ErrNameTaken)
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

	return userID, role, errors.New(errlist.ErrInvalidLogin)
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

	return user, errors.New(errlist.ErrUserNotFound)
}

func (p *UserData) OutAscendingBySpecialty(orderField string, specialty string) (users []UsersList) {
	for i := range p.fio {
		if (specialty == "" || p.specialty[i] == specialty) && p.role[i] == "teacher" {
			users = append(users, UsersList{
				ID:        p.id[i],
				Fio:       p.fio[i],
				Age:       p.age[i],
				Specialty: p.specialty[i],
				Price:     p.price[i],
				Rating:    p.rating[i],
			})
		}
	}

	switch orderField {
	case "price":
		sort.Slice(users, func(i, j int) bool {
			return users[i].Price < users[j].Price
		})
	case "rating":
		sort.Slice(users, func(i, j int) bool {
			return users[i].Rating < users[j].Rating
		})
	default:
	}

	return users
}

func (p *UserData) OutDescendingBySpecialty(orderField string, specialty string) (users []UsersList) {
	for i := range p.fio {
		if (specialty == "" || p.specialty[i] == specialty) && p.role[i] == "teacher" {
			users = append(users, UsersList{
				ID:        p.id[i],
				Fio:       p.fio[i],
				Age:       p.age[i],
				Specialty: p.specialty[i],
				Price:     p.price[i],
				Rating:    p.rating[i],
			})
		}
	}

	switch orderField {
	case "price":
		sort.Slice(users, func(i, j int) bool {
			return users[i].Price > users[j].Price
		})
	case "rating":
		sort.Slice(users, func(i, j int) bool {
			return users[i].Rating > users[j].Rating
		})
	default:
	}

	return users
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

func (p *UserData) AddRating(userID uuid.UUID, rating uint8) {
	for i, val := range p.id {
		if val == userID {
			p.rating[i] = (p.rating[i] + float32(rating)) / 2
		}
	}
}

func (p *UserData) StudentsByTeacher(teacherID uuid.UUID) (users []UsersList) {
	var students []uuid.UUID

	for i, val := range p.id {
		if val == teacherID {
			students = p.students[i]
			break
		}
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
				break
			}
		}
	}

	return users
}

func (p *UserData) EditGrade(studentID uuid.UUID, grade float32) {
	for i, val := range p.id {
		if val == studentID && p.role[i] == "student" {
			p.rating[i] = grade
			return
		}
	}
}

func (p *UserData) AddRequest(studentID uuid.UUID, teacherID uuid.UUID) {
	for i, val := range p.id {
		if val == teacherID && p.role[i] == "teacher" {
			p.requests[i] = append(p.requests[i], studentID)
			break
		}
	}

	for i, val := range p.id {
		if val == studentID {
			p.requests[i] = append(p.requests[i], teacherID)
			return
		}
	}

}

//func (p *UserData) Accept()

//func (p *UserData) Deny()

func (p *UserData) FillProfile(userID uuid.UUID, userData UsersList) {
	for i, val := range p.id {
		if val == userID {
			p.age[i] = userData.Age
			p.fio[i] = userData.Fio
			p.specialty[i] = userData.Specialty
			p.price[i] = userData.Price
		}
	}
}
