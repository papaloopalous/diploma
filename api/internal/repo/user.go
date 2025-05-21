package repo

import (
	"api/internal/proto/userpb"
	"context"
	"sort"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UserRepoGRPC реализует взаимодействие с сервисом пользователей через gRPC
type UserRepoGRPC struct {
	db userpb.UserServiceClient // gRPC клиент для взаимодействия с сервисом пользователей
}

// Проверка реализации интерфейса UserRepo
var _ UserRepo = &UserRepoGRPC{}

// NewUserRepo создает новый экземпляр репозитория пользователей
func NewUserRepo(conn *grpc.ClientConn) *UserRepoGRPC {
	return &UserRepoGRPC{
		db: userpb.NewUserServiceClient(conn),
	}
}

const (
	userToken = "user-token"
)

// CreateAccount создает новую учетную запись
func (r *UserRepoGRPC) CreateAccount(username string, pass string, role string) (uuid.UUID, error) {
	md := metadata.New(map[string]string{
		authorization: bearer + userToken,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	existsResp, err := r.db.UserExists(ctx, &userpb.UsernameRequest{Username: username})
	if err != nil {
		return uuid.Nil, err
	}
	if existsResp.Exists {
		return uuid.Nil, err
	}
	resp, err := r.db.AddUser(ctx, &userpb.NewUserRequest{
		Username: username,
		Password: pass,
		Role:     role,
	})
	if err != nil {
		return uuid.Nil, err
	}
	return uuid.MustParse(resp.Id), nil
}

// CheckPass проверяет учетные данные пользователя
func (r *UserRepoGRPC) CheckPass(username string, pass string) (uuid.UUID, string, error) {
	md := metadata.New(map[string]string{
		authorization: bearer + userToken,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	resp, err := r.db.CheckCredentials(ctx, &userpb.CredentialsRequest{
		Username: username,
		Password: pass,
	})
	if err != nil {
		return uuid.Nil, "", err
	}
	return uuid.MustParse(resp.Id), resp.Role, nil
}

// FindUser находит пользователя по ID
func (r *UserRepoGRPC) FindUser(userID uuid.UUID) (UsersList, error) {
	md := metadata.New(map[string]string{
		authorization: bearer + userToken,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	resp, err := r.db.GetUserByID(ctx, &userpb.UserIDRequest{Id: userID.String()})
	if err != nil {
		return UsersList{}, err
	}
	return UsersList{
		ID:        uuid.MustParse(resp.Id),
		Fio:       resp.Fio,
		Age:       uint8(resp.Age),
		Specialty: resp.Specialty,
		Price:     int(resp.Price),
		Rating:    resp.Rating,
	}, nil
}

// FillProfile обновляет профиль пользователя
func (r *UserRepoGRPC) FillProfile(userID uuid.UUID, userData UsersList) error {
	md := metadata.New(map[string]string{
		authorization: bearer + userToken,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	_, err := r.db.UpdateUserProfile(ctx, &userpb.UpdateProfileRequest{
		Id:        userID.String(),
		Fio:       userData.Fio,
		Age:       uint32(userData.Age),
		Specialty: userData.Specialty,
		Price:     int32(userData.Price),
	})
	if err != nil {
		return err
	}
	return nil
}

// AddRequest создает запрос на обучение
func (r *UserRepoGRPC) AddRequest(studentID, teacherID uuid.UUID) error {
	md := metadata.New(map[string]string{
		authorization: bearer + userToken,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	_, err := r.db.AddRequestLink(ctx, &userpb.RelationRequest{
		FromId: studentID.String(),
		ToId:   teacherID.String(),
	})
	if err != nil {
		return err
	}
	return nil
}

// Accept подтверждает запрос на обучение
func (r *UserRepoGRPC) Accept(teacherID, studentID uuid.UUID) error {
	md := metadata.New(map[string]string{
		authorization: bearer + userToken,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	_, err := r.db.AcceptRequest(ctx, &userpb.RelationRequest{
		FromId: teacherID.String(),
		ToId:   studentID.String(),
	})
	if err != nil {
		return err
	}
	return nil
}

// Deny отклоняет запрос на обучение
func (r *UserRepoGRPC) Deny(teacherID, studentID uuid.UUID) error {
	md := metadata.New(map[string]string{
		authorization: bearer + userToken,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	_, err := r.db.DenyRequest(ctx, &userpb.RelationRequest{
		FromId: teacherID.String(),
		ToId:   studentID.String(),
	})
	if err != nil {
		return err
	}
	return nil
}

// ShowRequests возвращает список запросов на обучение
func (r *UserRepoGRPC) ShowRequests(userID uuid.UUID) ([]UsersList, error) {
	md := metadata.New(map[string]string{
		authorization: bearer + userToken,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	reqIDs, err := r.db.GetRequests(ctx, &userpb.UserIDRequest{Id: userID.String()})
	if err != nil {
		return nil, err
	}
	resp, err := r.db.GetUsersByIDs(ctx, &userpb.UUIDListRequest{Ids: reqIDs.Ids})
	if err != nil {
		return nil, err
	}
	var users []UsersList
	for _, u := range resp.Users {
		users = append(users, UsersList{
			ID:        uuid.MustParse(u.Id),
			Fio:       u.Fio,
			Age:       uint8(u.Age),
			Specialty: u.Specialty,
			Price:     int(u.Price),
			Rating:    u.Rating,
		})
	}
	return users, nil
}

// AddRating добавляет и усредняет оценку преподавателя
func (r *UserRepoGRPC) AddRating(userID uuid.UUID, newRating float32) error {
	md := metadata.New(map[string]string{
		authorization: bearer + userToken,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	resp, err := r.db.GetRating(ctx, &userpb.UserIDRequest{Id: userID.String()})
	if err != nil {
		return err
	}

	newRating = (resp.Rating + newRating) / 2

	_, err = r.db.UpdateRating(ctx, &userpb.UpdateRatingRequest{
		Id:        userID.String(),
		NewRating: newRating,
	})
	return err
}

// HasThatTeacher проверяет связь студента с преподавателем
func (r *UserRepoGRPC) HasThatTeacher(studentID, teacherID uuid.UUID) (bool, error) {
	md := metadata.New(map[string]string{
		authorization: bearer + userToken,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	resp, err := r.db.HasTeacher(ctx, &userpb.RelationRequest{
		FromId: studentID.String(),
		ToId:   teacherID.String(),
	})
	if err != nil {
		return false, err
	}
	return resp.Result, nil
}

// StudentsByTeacher возвращает список студентов преподавателя
func (r *UserRepoGRPC) StudentsByTeacher(teacherID uuid.UUID) ([]UsersList, error) {
	md := metadata.New(map[string]string{
		authorization: bearer + userToken,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	resp, err := r.db.GetStudentsByTeacher(ctx, &userpb.UserIDRequest{Id: teacherID.String()})
	if err != nil {
		return nil, err
	}
	var users []UsersList
	for _, u := range resp.Users {
		users = append(users, UsersList{
			ID:     uuid.MustParse(u.Id),
			Fio:    u.Fio,
			Age:    uint8(u.Age),
			Rating: u.Rating,
		})
	}
	return users, nil
}

// TeachersByStudent возвращает список преподавателей студента
func (r *UserRepoGRPC) TeachersByStudent(studentID uuid.UUID) ([]UsersList, error) {
	md := metadata.New(map[string]string{
		authorization: bearer + userToken,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	resp, err := r.db.GetTeachersByStudent(ctx, &userpb.UserIDRequest{Id: studentID.String()})
	if err != nil {
		return nil, err
	}
	var teachers []UsersList
	for _, u := range resp.Users {
		teachers = append(teachers, UsersList{
			ID:        uuid.MustParse(u.Id),
			Fio:       u.Fio,
			Age:       uint8(u.Age),
			Specialty: u.Specialty,
			Price:     int(u.Price),
			Rating:    u.Rating,
		})
	}
	return teachers, nil
}

// EditGrade обновляет среднюю оценку студента
func (r *UserRepoGRPC) EditGrade(studentID uuid.UUID, grade float32) error {
	md := metadata.New(map[string]string{
		authorization: bearer + userToken,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	_, err := r.db.UpdateRating(ctx, &userpb.UpdateRatingRequest{
		Id:        studentID.String(),
		NewRating: grade,
	})
	if err != nil {
		return err
	}
	return nil
}

// outBySpecialty вспомогательная функция для сортировки преподавателей
func (r *UserRepoGRPC) outBySpecialty(orderField, specialty string, studentID uuid.UUID, ascending bool) ([]UsersList, error) {
	md := metadata.New(map[string]string{
		authorization: bearer + userToken,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	links, err := r.db.GetUserLinks(ctx, &userpb.UserIDRequest{Id: studentID.String()})
	if err != nil {
		return nil, err
	}

	exclude := append(links.Teachers, links.Requests...)

	resp, err := r.db.GetAvailableTeachers(ctx, &userpb.AvailableTeachersRequest{
		Specialty: specialty,
		Exclude:   exclude,
	})

	if err != nil {
		return nil, err
	}

	users := make([]UsersList, 0, len(resp.Users))
	for _, u := range resp.Users {
		users = append(users, UsersList{
			ID:        uuid.MustParse(u.Id),
			Fio:       u.Fio,
			Age:       uint8(u.Age),
			Specialty: u.Specialty,
			Price:     int(u.Price),
			Rating:    u.Rating,
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

	return users, nil
}

// OutAscendingBySpecialty возвращает отсортированный по возрастанию список преподавателей
func (r *UserRepoGRPC) OutAscendingBySpecialty(orderField, specialty string, studentID uuid.UUID) ([]UsersList, error) {
	return r.outBySpecialty(orderField, specialty, studentID, true)
}

// OutDescendingBySpecialty возвращает отсортированный по убыванию список преподавателей
func (r *UserRepoGRPC) OutDescendingBySpecialty(orderField, specialty string, studentID uuid.UUID) ([]UsersList, error) {
	return r.outBySpecialty(orderField, specialty, studentID, false)
}
