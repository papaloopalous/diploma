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
	id               []uuid.UUID
	name             []string
	student          []uuid.UUID
	studentFIO       []string
	teacher          []uuid.UUID
	teacherFIO       []string
	fileNameTask     []string
	fileDataTask     [][]byte
	fileNameSolution []string
	fileDataSolution [][]byte
	grade            []uint8
	status           []string
}

type TaskRepo interface {
	CreateTask(teacher uuid.UUID, student uuid.UUID, name string, studentFIO string, teacherFIO string) uuid.UUID
	GetTask(taskID uuid.UUID) (fileName string, fileData []byte, taskName string, err error)
	GetSolution(taskID uuid.UUID) (fileName string, fileData []byte, taskName string, err error)
	LinkFileTask(taskID uuid.UUID, fileName string, fileData []byte)
	LinkFileSolution(taskID uuid.UUID, fileName string, fileData []byte)
	Grade(taskID uuid.UUID, grade uint8) (studentID uuid.UUID)
	Solve(taskID uuid.UUID)
	AvgGrade(studentID uuid.UUID) (grade float32)
	AllTasks(userID uuid.UUID) (tasks []taskList)
}

type taskList struct {
	ID      uuid.UUID `json:"taskID"`
	Name    string    `json:"taskName"`
	Grade   uint8     `json:"grade,omitempty"`
	Status  string    `json:"status"`
	Student string    `json:"student"`
	Teacher string    `json:"teacher"`
}

var _ TaskRepo = &TaskData{}

func NewTaskRepo() *TaskData {
	return &TaskData{}
}

func (p *TaskData) CreateTask(teacher uuid.UUID, student uuid.UUID, name string, studentFIO string, teacherFIO string) uuid.UUID {
	id := uuid.New()
	p.id = append(p.id, id)
	p.student = append(p.student, student)
	p.teacher = append(p.teacher, teacher)
	p.name = append(p.name, name)
	p.status = append(p.status, statusSent)
	p.fileDataTask = append(p.fileDataTask, []byte{})
	p.fileNameTask = append(p.fileNameTask, "")
	p.fileDataSolution = append(p.fileDataSolution, []byte{})
	p.fileNameSolution = append(p.fileNameSolution, "")
	p.grade = append(p.grade, 0)
	p.studentFIO = append(p.studentFIO, studentFIO)
	p.teacherFIO = append(p.teacherFIO, teacherFIO)

	return id
}

func (p *TaskData) GetTask(taskID uuid.UUID) (fileName string, fileData []byte, taskName string, err error) {
	for i, val := range p.id {
		if val == taskID {
			fileName = p.fileNameTask[i]
			fileData = p.fileDataTask[i]
			taskName = p.name[i]
			return fileName, fileData, taskName, nil
		}
	}

	return fileName, fileData, taskName, errors.New(errlist.ErrNoTask)
}

func (p *TaskData) GetSolution(taskID uuid.UUID) (fileName string, fileData []byte, taskName string, err error) {
	for i, val := range p.id {
		if val == taskID {
			fileName = p.fileNameSolution[i]
			fileData = p.fileDataSolution[i]
			taskName = p.name[i]
			return fileName, fileData, taskName, nil
		}
	}

	return fileName, fileData, taskName, errors.New(errlist.ErrNoTask)
}

func (p *TaskData) LinkFileTask(taskID uuid.UUID, fileName string, fileData []byte) {
	for i, val := range p.id {
		if val == taskID {
			p.fileDataTask[i] = fileData
			p.fileNameTask[i] = fileName
			return
		}
	}
}

func (p *TaskData) LinkFileSolution(taskID uuid.UUID, fileName string, fileData []byte) {
	for i, val := range p.id {
		if val == taskID {
			p.fileDataSolution[i] = fileData
			p.fileNameSolution[i] = fileName
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
				Name:    p.name[i],
				Status:  p.status[i],
				ID:      val,
				Grade:   p.grade[i],
				Student: p.studentFIO[i],
				Teacher: p.teacherFIO[i],
			})
		}
	}

	return tasks
}
