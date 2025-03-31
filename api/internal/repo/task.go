package repo

import (
	errlist "api/internal/errList"
	"errors"

	"github.com/google/uuid"
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
	p.status = append(p.status, "sent to student")
	p.fileData = append(p.fileData, []byte{})
	p.fileName = append(p.fileName, "")

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
