package repo

import (
	"api/internal/proto/taskpb"
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

// TaskRepoGRPC реализует взаимодействие с сервисом заданий через gRPC
type TaskRepoGRPC struct {
	db taskpb.TaskServiceClient // gRPC клиент для взаимодействия с сервисом заданий
}

// Проверка реализации интерфейса TaskRepo
var _ TaskRepo = &TaskRepoGRPC{}

// NewTaskRepo создает новый экземпляр репозитория заданий
func NewTaskRepo(conn *grpc.ClientConn) *TaskRepoGRPC {
	return &TaskRepoGRPC{
		db: taskpb.NewTaskServiceClient(conn),
	}
}

// CreateTask создает новое задание в базе данных
func (r *TaskRepoGRPC) CreateTask(teacher uuid.UUID, student uuid.UUID, name string, studentFIO string, teacherFIO string) (uuid.UUID, error) {
	ctx := context.Background()
	resp, err := r.db.CreateTask(ctx, &taskpb.CreateTaskRequest{
		TeacherId:  teacher.String(),
		StudentId:  student.String(),
		Name:       name,
		StudentFio: studentFIO,
		TeacherFio: teacherFIO,
	})
	if err != nil {
		return uuid.Nil, err
	}
	return uuid.MustParse(resp.Id), err
}

// GetTask получает файл задания из хранилища
func (r *TaskRepoGRPC) GetTask(taskID uuid.UUID) (fileName string, fileData []byte, err error) {
	ctx := context.Background()
	resp, err := r.db.GetTask(ctx, &taskpb.TaskIDRequest{
		Id: taskID.String(),
	})
	if err != nil {
		return "", nil, err
	}
	if len(resp.FileData) == 0 {
		return "", nil, err
	}
	return resp.FileName, resp.FileData, nil
}

// GetSolution получает файл решения из хранилища
func (r *TaskRepoGRPC) GetSolution(taskID uuid.UUID) (fileName string, fileData []byte, err error) {
	ctx := context.Background()
	resp, err := r.db.GetSolution(ctx, &taskpb.TaskIDRequest{
		Id: taskID.String(),
	})
	if err != nil {
		return "", nil, err
	}
	if len(resp.FileData) == 0 {
		return "", nil, err
	}
	return resp.FileName, resp.FileData, nil
}

// LinkFileTask прикрепляет файл к заданию в хранилище
func (r *TaskRepoGRPC) LinkFileTask(taskID uuid.UUID, fileName string, fileData []byte) error {
	ctx := context.Background()
	_, err := r.db.LinkFileTask(ctx, &taskpb.LinkFileRequest{
		TaskId:   taskID.String(),
		FileName: fileName,
		FileData: fileData,
	})
	if err != nil {
		return err
	}
	return nil
}

// LinkFileSolution прикрепляет файл решения к заданию в хранилище
func (r *TaskRepoGRPC) LinkFileSolution(taskID uuid.UUID, fileName string, fileData []byte) error {
	ctx := context.Background()
	_, err := r.db.LinkFileSolution(ctx, &taskpb.LinkFileRequest{
		TaskId:   taskID.String(),
		FileName: fileName,
		FileData: fileData,
	})
	if err != nil {
		return err
	}
	return nil
}

// Grade выставляет оценку за задание и возвращает ID студента
func (r *TaskRepoGRPC) Grade(taskID uuid.UUID, grade uint8) (studentID uuid.UUID, err error) {
	ctx := context.Background()
	resp, err := r.db.Grade(ctx, &taskpb.GradeRequest{
		TaskId: taskID.String(),
		Grade:  uint32(grade),
	})
	if err != nil {
		return uuid.Nil, err
	}
	return uuid.MustParse(resp.StudentId), nil
}

// Solve отмечает задание как решенное
func (r *TaskRepoGRPC) Solve(taskID uuid.UUID) error {
	ctx := context.Background()
	_, err := r.db.Solve(ctx, &taskpb.TaskIDRequest{
		Id: taskID.String(),
	})
	if err != nil {
		return err
	}
	return nil
}

// AvgGrade вычисляет среднюю оценку студента по всем заданиям
func (r *TaskRepoGRPC) AvgGrade(studentID uuid.UUID) (grade float32, err error) {
	ctx := context.Background()
	resp, err := r.db.AvgGrade(ctx, &taskpb.StudentIDRequest{
		StudentId: studentID.String(),
	})
	if err != nil {
		return 0, err
	}
	return resp.Grade, nil
}

// AllTasks возвращает список всех заданий пользователя
func (r *TaskRepoGRPC) AllTasks(userID uuid.UUID) []taskList {
	ctx := context.Background()
	resp, err := r.db.AllTasks(ctx, &taskpb.UserIDRequest{
		UserId: userID.String(),
	})
	if err != nil {
		return nil
	}

	tasks := make([]taskList, 0, len(resp.Tasks))
	for _, t := range resp.Tasks {
		tasks = append(tasks, taskList{
			Name:    t.Name,
			Status:  t.Status,
			ID:      uuid.MustParse(t.Id),
			Grade:   uint8(t.Grade),
			Student: t.Student,
			Teacher: t.Teacher,
		})
	}
	return tasks
}
