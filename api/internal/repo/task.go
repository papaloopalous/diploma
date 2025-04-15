package repo

import (
	"api/internal/messages"
	"errors"

	"github.com/google/uuid"
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

func (p *TaskData) GetTask(taskID uuid.UUID) (fileName string, fileData []byte, err error) {
	for i, val := range p.id {
		if val == taskID {
			fileData = p.fileDataTask[i]
			if fileData == nil {
				return fileName, fileData, errors.New(messages.ErrTaskEmpty)
			}
			fileName = p.fileNameTask[i]
			return fileName, fileData, nil
		}
	}

	return fileName, fileData, errors.New(messages.ErrTaskNotFound)
}

func (p *TaskData) GetSolution(taskID uuid.UUID) (fileName string, fileData []byte, err error) {
	for i, val := range p.id {
		if val == taskID {
			fileData = p.fileDataSolution[i]
			if fileData == nil {
				return fileName, fileData, errors.New(messages.ErrSolutionEmpty)
			}
			fileName = p.fileNameSolution[i]
			return fileName, fileData, nil
		}
	}

	return fileName, fileData, errors.New(messages.ErrTaskNotFound)
}

func (p *TaskData) LinkFileTask(taskID uuid.UUID, fileName string, fileData []byte) error {
	for i, val := range p.id {
		if val == taskID {
			p.fileDataTask[i] = fileData
			p.fileNameTask[i] = fileName
			return nil
		}
	}

	return errors.New(messages.ErrTaskNotFound)
}

func (p *TaskData) LinkFileSolution(taskID uuid.UUID, fileName string, fileData []byte) error {
	for i, val := range p.id {
		if val == taskID {
			p.fileDataSolution[i] = fileData
			p.fileNameSolution[i] = fileName
			return nil
		}
	}

	return errors.New(messages.ErrTaskNotFound)
}

func (p *TaskData) Grade(taskID uuid.UUID, grade uint8) (studentID uuid.UUID, err error) {
	for i, val := range p.id {
		if val == taskID {
			p.grade[i] = grade
			p.status[i] = statusGraded
			studentID = p.student[i]
			return studentID, nil
		}
	}

	return studentID, errors.New(messages.ErrTaskNotFound)
}

func (p *TaskData) Solve(taskID uuid.UUID) error {
	for i, val := range p.id {
		if val == taskID {
			p.status[i] = statusSolved
			return nil
		}
	}

	return errors.New(messages.ErrTaskNotFound)
}

func (p *TaskData) AvgGrade(studentID uuid.UUID) (grade float32, err error) {
	var count float32 = 0

	found := false
	for i, val := range p.student {
		if val == studentID {
			found = true
			if p.status[i] == statusGraded {
				grade += float32(p.grade[i])
				count++
				found = true
			}
		}
	}

	if !found {
		return grade, errors.New(messages.ErrStudentNotFound)
	}

	if count != 0 {
		return grade / count, nil
	}

	return grade, nil
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
