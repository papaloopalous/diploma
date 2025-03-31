package repo

import (
	errlist "api/internal/errList"
	"errors"

	"github.com/google/uuid"
)

const (
	statusSent   = "sent to student"
	statusSolved = "ready to grade"
	statusGraded = "graded"
)

type TaskData struct {
	id       []uuid.UUID
	name     []string
	student  []uuid.UUID
	teacher  []uuid.UUID
	fileName []string
	fileData [][]byte
	grade    []uint8
	status   []string
}

type TaskRepo interface {
	CreateTask(teacher uuid.UUID, student uuid.UUID, name string) uuid.UUID
	GetTask(taskID uuid.UUID) (fileName string, fileData []byte, taskName string, err error)
	LinkFile(taskID uuid.UUID, fileName string, fileData []byte)
	Grade(taskID uuid.UUID, grade uint8) (studentID uuid.UUID)
	Solve(taskID uuid.UUID)
	AvgGrade(studentID uuid.UUID) (grade float32)
	AllTasks(userID uuid.UUID) (tasks []taskList)
}

type taskList struct {
	ID     uuid.UUID `json:"taskID"`
	Name   string    `json:"taskName"`
	Grade  uint8     `json:"grade,omitempty"`
	Status string    `json:"status"`
}

var _ TaskRepo = &TaskData{}

func NewTaskRepo() *TaskData {
	return &TaskData{}
}

func (p *TaskData) CreateTask(teacher uuid.UUID, student uuid.UUID, name string) uuid.UUID {
	id := uuid.New()
	p.id = append(p.id, id)
	p.student = append(p.student, student)
	p.teacher = append(p.teacher, teacher)
	p.name = append(p.name, name)
	p.status = append(p.status, statusSent)
	p.fileData = append(p.fileData, []byte{})
	p.fileName = append(p.fileName, "")
	p.grade = append(p.grade, 0)

	return id
}

func (p *TaskData) GetTask(taskID uuid.UUID) (fileName string, fileData []byte, taskName string, err error) {
	for i, val := range p.id {
		if val == taskID {
			fileName = p.fileName[i]
			fileData = p.fileData[i]
			taskName = p.name[i]
			return fileName, fileData, taskName, nil
		}
	}

	return fileName, fileData, taskName, errors.New(errlist.ErrNoTask)
}

func (p *TaskData) LinkFile(taskID uuid.UUID, fileName string, fileData []byte) {
	for i, val := range p.id {
		if val == taskID {
			p.fileData[i] = fileData
			p.fileName[i] = fileName
			return
		}
	}
}

func (p *TaskData) Grade(taskID uuid.UUID, grade uint8) (studentID uuid.UUID) {
	for i, val := range p.id {
		if val == taskID {
			p.grade[i] = grade
			p.status[i] = statusGraded
			studentID = p.student[i]
			return
		}
	}

	return studentID
}

func (p *TaskData) Solve(taskID uuid.UUID) {
	for i, val := range p.id {
		if val == taskID {
			p.status[i] = statusSolved
			return
		}
	}
}

func (p *TaskData) AvgGrade(studentID uuid.UUID) (grade float32) {
	var count float32 = 0

	for i, val := range p.student {
		if val == studentID && p.status[i] == statusGraded {
			grade += float32(p.grade[i])
			count++
		}
	}

	if count != 0 {
		return grade / count
	}

	return 0
}

func (p *TaskData) AllTasks(userID uuid.UUID) (tasks []taskList) {
	for i, val := range p.id {
		if p.student[i] == userID || p.teacher[i] == userID {
			tasks = append(tasks, taskList{
				Name:   p.name[i],
				Status: p.status[i],
				ID:     val,
				Grade:  p.grade[i],
			})
		}
	}

	return tasks
}
