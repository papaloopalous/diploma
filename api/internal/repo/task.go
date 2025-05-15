package repo

import (
	"api/internal/messages"
	"api/internal/proto/taskpb"
	"context"
	"errors"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type TaskRepoGRPC struct {
	db taskpb.TaskServiceClient
}

var _ TaskRepo = &TaskRepoGRPC{}

func NewTaskRepo(conn *grpc.ClientConn) *TaskRepoGRPC {
	return &TaskRepoGRPC{
		db: taskpb.NewTaskServiceClient(conn),
	}
}

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

func (r *TaskRepoGRPC) GetTask(taskID uuid.UUID) (fileName string, fileData []byte, err error) {
	ctx := context.Background()
	resp, err := r.db.GetTask(ctx, &taskpb.TaskIDRequest{
		Id: taskID.String(),
	})
	if err != nil {
		return "", nil, errors.New(messages.ErrTaskNotFound)
	}
	if len(resp.FileData) == 0 {
		return "", nil, errors.New(messages.ErrTaskEmpty)
	}
	return resp.FileName, resp.FileData, nil
}

func (r *TaskRepoGRPC) GetSolution(taskID uuid.UUID) (fileName string, fileData []byte, err error) {
	ctx := context.Background()
	resp, err := r.db.GetSolution(ctx, &taskpb.TaskIDRequest{
		Id: taskID.String(),
	})
	if err != nil {
		return "", nil, errors.New(messages.ErrTaskNotFound)
	}
	if len(resp.FileData) == 0 {
		return "", nil, errors.New(messages.ErrSolutionEmpty)
	}
	return resp.FileName, resp.FileData, nil
}

func (r *TaskRepoGRPC) LinkFileTask(taskID uuid.UUID, fileName string, fileData []byte) error {
	ctx := context.Background()
	_, err := r.db.LinkFileTask(ctx, &taskpb.LinkFileRequest{
		TaskId:   taskID.String(),
		FileName: fileName,
		FileData: fileData,
	})
	if err != nil {
		return errors.New(messages.ErrTaskNotFound)
	}
	return nil
}

func (r *TaskRepoGRPC) LinkFileSolution(taskID uuid.UUID, fileName string, fileData []byte) error {
	ctx := context.Background()
	_, err := r.db.LinkFileSolution(ctx, &taskpb.LinkFileRequest{
		TaskId:   taskID.String(),
		FileName: fileName,
		FileData: fileData,
	})
	if err != nil {
		return errors.New(messages.ErrTaskNotFound)
	}
	return nil
}

func (r *TaskRepoGRPC) Grade(taskID uuid.UUID, grade uint8) (studentID uuid.UUID, err error) {
	ctx := context.Background()
	resp, err := r.db.Grade(ctx, &taskpb.GradeRequest{
		TaskId: taskID.String(),
		Grade:  uint32(grade),
	})
	if err != nil {
		return uuid.Nil, errors.New(messages.ErrTaskNotFound)
	}
	return uuid.MustParse(resp.StudentId), nil
}

func (r *TaskRepoGRPC) Solve(taskID uuid.UUID) error {
	ctx := context.Background()
	_, err := r.db.Solve(ctx, &taskpb.TaskIDRequest{
		Id: taskID.String(),
	})
	if err != nil {
		return errors.New(messages.ErrTaskNotFound)
	}
	return nil
}

func (r *TaskRepoGRPC) AvgGrade(studentID uuid.UUID) (grade float32, err error) {
	ctx := context.Background()
	resp, err := r.db.AvgGrade(ctx, &taskpb.StudentIDRequest{
		StudentId: studentID.String(),
	})
	if err != nil {
		return 0, errors.New(messages.ErrStudentNotFound)
	}
	return resp.Grade, nil
}

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
