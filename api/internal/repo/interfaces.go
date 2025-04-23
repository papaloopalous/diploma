package repo

import "github.com/google/uuid"

//users

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
	OutAscendingBySpecialty(orderField string, specialty string, userID uuid.UUID) (users []UsersList, err error)
	OutDescendingBySpecialty(orderField string, specialty string, userID uuid.UUID) (users []UsersList, err error)
	HasThatTeacher(studentID uuid.UUID, teacherID uuid.UUID) (bool, error)
	AddRating(userID uuid.UUID, rating float32) error
	StudentsByTeacher(teacherID uuid.UUID) (users []UsersList, err error)
	EditGrade(studentID uuid.UUID, grade float32) error
	FillProfile(userID uuid.UUID, userData UsersList) error
	TeachersByStudent(studentID uuid.UUID) (teachers []UsersList, err error)
	AddRequest(studentID uuid.UUID, teacherID uuid.UUID) error
	ShowRequests(userID uuid.UUID) (users []UsersList, err error)
	Accept(teacherID uuid.UUID, studentID uuid.UUID) error
	Deny(teacherID uuid.UUID, studentID uuid.UUID) error
}

//tasks

type taskList struct {
	ID      uuid.UUID `json:"taskID"`
	Name    string    `json:"taskName"`
	Grade   uint8     `json:"grade,omitempty"`
	Status  string    `json:"status"`
	Student string    `json:"student"`
	Teacher string    `json:"teacher"`
}

const (
	statusSent   = "sent to student"
	statusSolved = "ready to grade"
	statusGraded = "graded"
)

type TaskRepo interface {
	CreateTask(teacher uuid.UUID, student uuid.UUID, name string, studentFIO string, teacherFIO string) uuid.UUID
	GetTask(taskID uuid.UUID) (fileName string, fileData []byte, err error)
	GetSolution(taskID uuid.UUID) (fileName string, fileData []byte, err error)
	LinkFileTask(taskID uuid.UUID, fileName string, fileData []byte) error
	LinkFileSolution(taskID uuid.UUID, fileName string, fileData []byte) error
	Grade(taskID uuid.UUID, grade uint8) (studentID uuid.UUID, err error)
	Solve(taskID uuid.UUID) error
	AvgGrade(studentID uuid.UUID) (grade float32, err error)
	AllTasks(userID uuid.UUID) (tasks []taskList)
}
