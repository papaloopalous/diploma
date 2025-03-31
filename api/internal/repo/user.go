package repo

import (
	errlist "api/internal/errList"
	"errors"

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
}

type UserRepo interface {
	FindUser(userID uuid.UUID) (fio string, role string, age uint8, specialty string, err error)
	AddUser(fio string, username string, pass string, role string, age uint8, specialty string) (err error)
	CheckPass(username string, pass string) (userID uuid.UUID, role string, err error)
	CreateAccount(username string, pass string, role string) (userID uuid.UUID, err error)
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

func (p *UserData) AddUser(fio string, username string, pass string, role string, age uint8, specialty string) (err error) {
	for _, val := range p.username {
		if val == username {
			return errors.New(errlist.ErrNameTaken)
		}
	}

	p.fio = append(p.fio, fio)
	p.role = append(p.role, role)
	p.age = append(p.age, age)
	p.specialty = append(p.specialty, specialty)
	return nil
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

func (p *UserData) FindUser(userID uuid.UUID) (fio string, role string, age uint8, specialty string, err error) {
	for i, val := range p.id {
		if val == userID {
			fio = p.fio[i]
			role = p.role[i]
			age = p.age[i]
			specialty = p.specialty[i]
			return fio, role, age, specialty, nil
		}
	}

	return fio, role, age, specialty, errors.New(errlist.ErrUserNotFound)
}
