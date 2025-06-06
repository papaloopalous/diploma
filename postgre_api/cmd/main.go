package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"postgre_api/chatpb"
	"postgre_api/taskpb"
	"postgre_api/userpb"
	"strings"
	"time"

	"github.com/google/uuid"
	pgx "github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type server struct {
	userpb.UnimplementedUserServiceServer
	taskpb.UnimplementedTaskServiceServer
	chatpb.UnimplementedChatServiceServer
	db *pgx.Conn
}

func (s *server) AddUser(ctx context.Context, req *userpb.NewUserRequest) (*userpb.UserIDResponse, error) {
	id := uuid.New()
	_, err := s.db.Exec(ctx, `
		INSERT INTO users (id, username, pass, role) 
		VALUES ($1, $2, $3, $4)
	`, id, req.Username, req.Password, req.Role)
	if err != nil {
		return nil, err
	}
	return &userpb.UserIDResponse{Id: id.String()}, nil
}

func (s *server) GetUserLinks(ctx context.Context, req *userpb.UserIDRequest) (*userpb.UserLinksResponse, error) {
	var teachers, requests []uuid.UUID

	err := s.db.QueryRow(ctx, `
        SELECT students FROM users WHERE id = $1
    `, req.Id).Scan(&teachers)
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(ctx, `
        SELECT requests FROM users WHERE id = $1
    `, req.Id).Scan(&requests)
	if err != nil {
		return nil, err
	}

	var teacherIDs, requestIDs []string
	for _, id := range teachers {
		teacherIDs = append(teacherIDs, id.String())
	}
	for _, id := range requests {
		requestIDs = append(requestIDs, id.String())
	}

	return &userpb.UserLinksResponse{
		Teachers: teacherIDs,
		Requests: requestIDs,
	}, nil
}

func (s *server) CheckCredentials(ctx context.Context, req *userpb.CredentialsRequest) (*userpb.CredentialsResponse, error) {
	var id uuid.UUID
	var role string
	err := s.db.QueryRow(ctx, `
		SELECT id, role FROM users 
		WHERE username = $1 AND pass = $2
	`, req.Username, req.Password).Scan(&id, &role)
	if err != nil {
		return nil, err
	}
	return &userpb.CredentialsResponse{Id: id.String(), Role: role}, nil
}

func (s *server) GetUserByID(ctx context.Context, req *userpb.UserIDRequest) (*userpb.UserProfileResponse, error) {
	var fio, specialty string
	var age int16
	var price int
	var rating sql.NullFloat64

	err := s.db.QueryRow(ctx, `
        SELECT fio, age, specialty, price, rating 
        FROM users WHERE id = $1
    `, req.Id).Scan(&fio, &age, &specialty, &price, &rating)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user with ID %s not found", req.Id)
		}
		return nil, err
	}

	return &userpb.UserProfileResponse{
		Id:        req.Id,
		Fio:       fio,
		Age:       uint32(age),
		Specialty: specialty,
		Price:     int32(price),
		Rating:    float32(rating.Float64),
	}, nil
}

func (s *server) UserExists(ctx context.Context, req *userpb.UsernameRequest) (*userpb.UserExistsResponse, error) {
	var exists bool
	err := s.db.QueryRow(ctx, `
        SELECT EXISTS (
            SELECT 1 FROM users WHERE username = $1
        )
    `, req.Username).Scan(&exists)
	if err != nil {
		return nil, err
	}
	return &userpb.UserExistsResponse{Exists: exists}, nil
}

func (s *server) UpdateUserProfile(ctx context.Context, req *userpb.UpdateProfileRequest) (*userpb.Empty, error) {
	_, err := s.db.Exec(ctx, `
		UPDATE users SET fio = $1, age = $2, specialty = $3, price = $4 
		WHERE id = $5
	`, req.Fio, req.Age, req.Specialty, req.Price, req.Id)
	if err != nil {
		return nil, err
	}
	return &userpb.Empty{}, nil
}

func (s *server) UpdateRating(ctx context.Context, req *userpb.UpdateRatingRequest) (*userpb.Empty, error) {
	_, err := s.db.Exec(ctx, `
        UPDATE users 
        SET rating = $1 
        WHERE id = $2
    `, req.NewRating, req.Id)
	if err != nil {
		return nil, err
	}

	return &userpb.Empty{}, nil
}

func (s *server) AcceptRequest(ctx context.Context, req *userpb.RelationRequest) (*userpb.Empty, error) {
	_, err := s.db.Exec(ctx, `
        UPDATE users SET students = array_append(students, $1) 
        WHERE id = $2
    `, req.FromId, req.ToId)
	if err != nil {
		return nil, err
	}

	_, err = s.db.Exec(ctx, `
        UPDATE users SET teachers = array_append(teachers, $2) 
        WHERE id = $1
    `, req.FromId, req.ToId)
	if err != nil {
		return nil, err
	}

	_, err = s.db.Exec(ctx, `
		UPDATE users SET requests = array_remove(requests, $1) 
		WHERE id = $2
	`, req.FromId, req.ToId)
	if err != nil {
		return nil, err
	}

	_, err = s.db.Exec(ctx, `
		UPDATE users SET requests = array_remove(requests, $2) 
		WHERE id = $1
	`, req.FromId, req.ToId)
	if err != nil {
		return nil, err
	}

	return &userpb.Empty{}, nil
}

func (s *server) HasTeacher(ctx context.Context, req *userpb.RelationRequest) (*userpb.BoolResponse, error) {
	var exists bool
	err := s.db.QueryRow(ctx, `
        SELECT COALESCE($1 = ANY(students), false)
        FROM users
        WHERE id = $2
    `, req.ToId, req.FromId).Scan(&exists)
	if err != nil {
		return nil, err
	}

	return &userpb.BoolResponse{Result: exists}, nil
}

func (s *server) AddRequestLink(ctx context.Context, req *userpb.RelationRequest) (*userpb.Empty, error) {
	_, err := s.db.Exec(ctx, `
		UPDATE users SET requests = array_append(requests, $1) 
		WHERE id = $2
	`, req.FromId, req.ToId)
	if err != nil {
		return nil, err
	}
	_, err = s.db.Exec(ctx, `
		UPDATE users SET requests = array_append(requests, $2) 
		WHERE id = $1
	`, req.FromId, req.ToId)
	if err != nil {
		return nil, err
	}
	return &userpb.Empty{}, nil
}

func (s *server) DenyRequest(ctx context.Context, req *userpb.RelationRequest) (*userpb.Empty, error) {
	_, err := s.db.Exec(ctx, `
		UPDATE users SET requests = array_remove(requests, $1) 
		WHERE id = $2
	`, req.FromId, req.ToId)
	if err != nil {
		return nil, err
	}

	_, err = s.db.Exec(ctx, `
		UPDATE users SET requests = array_remove(requests, $2) 
		WHERE id = $1
	`, req.FromId, req.ToId)
	if err != nil {
		return nil, err
	}
	return &userpb.Empty{}, nil
}

func (s *server) GetAvailableTeachers(ctx context.Context, req *userpb.AvailableTeachersRequest) (*userpb.UsersListResponse, error) {
	if len(req.Exclude) == 0 {
		req.Exclude = []string{}
	}

	rows, err := s.db.Query(ctx, `
        SELECT id, fio, age, specialty, price, rating
        FROM users 
        WHERE role = 'teacher' AND ($1 = '' OR specialty = $1) AND id != ALL($2)
    `, req.Specialty, req.Exclude)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*userpb.UserProfileResponse
	for rows.Next() {
		var user userpb.UserProfileResponse
		var rating sql.NullFloat64
		err := rows.Scan(&user.Id, &user.Fio, &user.Age, &user.Specialty, &user.Price, &rating)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	return &userpb.UsersListResponse{Users: users}, nil
}

func (s *server) GetRating(ctx context.Context, req *userpb.UserIDRequest) (*userpb.RatingResponse, error) {
	var rating float32
	err := s.db.QueryRow(ctx, `
		SELECT rating FROM users WHERE id = $1
	`, req.Id).Scan(&rating)
	if err != nil {
		return nil, err
	}
	return &userpb.RatingResponse{Rating: rating}, nil
}

func (s *server) GetRequests(ctx context.Context, req *userpb.UserIDRequest) (*userpb.UUIDListResponse, error) {
	var requests []uuid.UUID
	err := s.db.QueryRow(ctx, `
		SELECT requests FROM users WHERE id = $1
	`, req.Id).Scan(&requests)
	if err != nil {
		return nil, err
	}

	var ids []string
	for _, id := range requests {
		ids = append(ids, id.String())
	}
	return &userpb.UUIDListResponse{Ids: ids}, nil
}

func (s *server) GetStudentTeacherLinks(ctx context.Context, req *userpb.UserIDRequest) (*userpb.StudentTeacherLinksResponse, error) {
	var teachers []uuid.UUID
	err := s.db.QueryRow(ctx, `
		SELECT teachers FROM users WHERE id = $1
	`, req.Id).Scan(&teachers)
	if err != nil {
		return nil, err
	}

	var teacherIDs []string
	for _, id := range teachers {
		teacherIDs = append(teacherIDs, id.String())
	}
	return &userpb.StudentTeacherLinksResponse{Teachers: teacherIDs}, nil
}

func (s *server) GetStudentsByTeacher(ctx context.Context, req *userpb.UserIDRequest) (*userpb.UsersListResponse, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, fio, age, rating 
		FROM users 
		WHERE $1 = ANY(students)
	`, req.Id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*userpb.UserProfileResponse
	for rows.Next() {
		var user userpb.UserProfileResponse
		err := rows.Scan(&user.Id, &user.Fio, &user.Age, &user.Rating)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	return &userpb.UsersListResponse{Users: users}, nil
}

func (s *server) GetTeachersByStudent(ctx context.Context, req *userpb.UserIDRequest) (*userpb.UsersListResponse, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, fio, age, specialty, price, rating 
		FROM users 
		WHERE $1 = ANY(teachers)
	`, req.Id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*userpb.UserProfileResponse
	for rows.Next() {
		var user userpb.UserProfileResponse
		err := rows.Scan(&user.Id, &user.Fio, &user.Age, &user.Specialty, &user.Price, &user.Rating)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	return &userpb.UsersListResponse{Users: users}, nil
}

func (s *server) GetUsersByIDs(ctx context.Context, req *userpb.UUIDListRequest) (*userpb.UsersListResponse, error) {
	rows, err := s.db.Query(ctx, `
        SELECT id, fio, age, specialty, price, rating 
        FROM users 
        WHERE id = ANY($1)
    `, req.Ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*userpb.UserProfileResponse
	for rows.Next() {
		var user userpb.UserProfileResponse
		err := rows.Scan(&user.Id, &user.Fio, &user.Age, &user.Specialty, &user.Price, &user.Rating)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	return &userpb.UsersListResponse{Users: users}, nil
}

func (s *server) CreateTask(ctx context.Context, req *taskpb.CreateTaskRequest) (*taskpb.TaskIDResponse, error) {
	taskID := uuid.New()

	_, err := s.db.Exec(ctx, `
        INSERT INTO tasks (id, teacher_id, student_id, name, student_fio, teacher_fio, status) 
        VALUES ($1, $2, $3, $4, $5, $6, 'sent to student')
    `, taskID, req.TeacherId, req.StudentId, req.Name, req.StudentFio, req.TeacherFio)
	if err != nil {
		return nil, err
	}

	return &taskpb.TaskIDResponse{Id: taskID.String()}, nil
}

func (s *server) GetTask(ctx context.Context, req *taskpb.TaskIDRequest) (*taskpb.FileResponse, error) {
	var fileName string
	var fileData []byte

	err := s.db.QueryRow(ctx, `
        SELECT file_name_task, file_data_task 
        FROM tasks 
        WHERE id = $1
    `, req.Id).Scan(&fileName, &fileData)
	if err != nil {
		return nil, err
	}

	return &taskpb.FileResponse{
		FileName: fileName,
		FileData: fileData,
	}, nil
}

func (s *server) GetSolution(ctx context.Context, req *taskpb.TaskIDRequest) (*taskpb.FileResponse, error) {
	var fileName string
	var fileData []byte

	err := s.db.QueryRow(ctx, `
        SELECT file_name_solution, file_data_solution 
        FROM tasks 
        WHERE id = $1
    `, req.Id).Scan(&fileName, &fileData)
	if err != nil {
		return nil, err
	}

	return &taskpb.FileResponse{
		FileName: fileName,
		FileData: fileData,
	}, nil
}

func (s *server) LinkFileTask(ctx context.Context, req *taskpb.LinkFileRequest) (*taskpb.Empty, error) {
	_, err := s.db.Exec(ctx, `
        UPDATE tasks 
        SET file_name_task = $1, file_data_task = $2 
        WHERE id = $3
    `, req.FileName, req.FileData, req.TaskId)
	if err != nil {
		return nil, err
	}

	return &taskpb.Empty{}, nil
}

func (s *server) LinkFileSolution(ctx context.Context, req *taskpb.LinkFileRequest) (*taskpb.Empty, error) {
	_, err := s.db.Exec(ctx, `
        UPDATE tasks 
        SET file_name_solution = $1, file_data_solution = $2 
        WHERE id = $3
    `, req.FileName, req.FileData, req.TaskId)
	if err != nil {
		return nil, err
	}

	return &taskpb.Empty{}, nil
}

func (s *server) Grade(ctx context.Context, req *taskpb.GradeRequest) (*taskpb.StudentIDResponse, error) {
	var studentID string
	err := s.db.QueryRow(ctx, `
        UPDATE tasks 
        SET grade = $1, status = 'graded' 
        WHERE id = $2 
        RETURNING student_id
    `, req.Grade, req.TaskId).Scan(&studentID)
	if err != nil {
		return nil, err
	}

	return &taskpb.StudentIDResponse{StudentId: studentID}, nil
}

func (s *server) Solve(ctx context.Context, req *taskpb.TaskIDRequest) (*taskpb.Empty, error) {
	_, err := s.db.Exec(ctx, `
        UPDATE tasks 
        SET status = 'ready to grade' 
        WHERE id = $1
    `, req.Id)
	if err != nil {
		return nil, err
	}

	return &taskpb.Empty{}, nil
}

func (s *server) AvgGrade(ctx context.Context, req *taskpb.StudentIDRequest) (*taskpb.GradeResponse, error) {
	var avgGrade float32
	err := s.db.QueryRow(ctx, `
        SELECT COALESCE(AVG(grade::float), 0) 
        FROM tasks 
        WHERE student_id = $1 AND status = 'graded'
    `, req.StudentId).Scan(&avgGrade)
	if err != nil {
		return nil, err
	}

	return &taskpb.GradeResponse{Grade: avgGrade}, nil
}

func (s *server) AllTasks(ctx context.Context, req *taskpb.UserIDRequest) (*taskpb.TaskListResponse, error) {
	rows, err := s.db.Query(ctx, `
        SELECT id, name, status, COALESCE(grade, 0), student_fio, teacher_fio 
        FROM tasks 
        WHERE student_id = $1 OR teacher_id = $1
    `, req.UserId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*taskpb.TaskInfo
	for rows.Next() {
		task := &taskpb.TaskInfo{}
		err := rows.Scan(&task.Id, &task.Name, &task.Status, &task.Grade, &task.Teacher, &task.Student)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return &taskpb.TaskListResponse{Tasks: tasks}, nil
}

func (s *server) CreateRoom(ctx context.Context,
	req *chatpb.CreateRoomRequest) (*chatpb.CreateRoomResponse, error) {

	var roomID uuid.UUID
	err := s.db.QueryRow(ctx, `
		SELECT id FROM chat_rooms
		WHERE (user1_id = $1 AND user2_id = $2)
		   OR (user1_id = $2 AND user2_id = $1)
	`, req.User1Id, req.User2Id).Scan(&roomID)

	switch err {
	case nil:
		return &chatpb.CreateRoomResponse{
			RoomId:        roomID.String(),
			AlreadyExists: true,
		}, nil
	case pgx.ErrNoRows:
		roomID = uuid.New()
		_, err = s.db.Exec(ctx, `
			INSERT INTO chat_rooms (id, user1_id, user2_id)
			VALUES ($1, $2, $3)
		`, roomID, req.User1Id, req.User2Id)
		if err != nil {
			return nil, err
		}
		return &chatpb.CreateRoomResponse{
			RoomId:        roomID.String(),
			AlreadyExists: false,
		}, nil
	default:
		return nil, err
	}
}

func (s *server) History(ctx context.Context,
	req *chatpb.RoomIDRequest) (*chatpb.HistoryResponse, error) {

	rows, err := s.db.Query(ctx, `
		SELECT id, sender_id, text, sent_at, status
		FROM chat_messages
		WHERE room_id = $1
		ORDER BY sent_at
	`, req.RoomId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []*chatpb.MessageInfo
	for rows.Next() {
		var (
			id, senderID uuid.UUID
			text         string
			sentAt       time.Time
			status       int16
		)
		if err = rows.Scan(&id, &senderID, &text, &sentAt, &status); err != nil {
			return nil, err
		}
		msgs = append(msgs, &chatpb.MessageInfo{
			Id:       id.String(),
			RoomId:   req.RoomId,
			SenderId: senderID.String(),
			Text:     text,
			SentAt:   timestamppb.New(sentAt),
			Status:   chatpb.MessageStatus(status),
		})
	}
	return &chatpb.HistoryResponse{Messages: msgs}, nil
}

func (s *server) SendMessage(ctx context.Context,
	req *chatpb.SendMessageRequest) (*chatpb.Empty, error) {

	m := req.GetMessage()
	if m.Id == "" {
		m.Id = uuid.New().String()
	}
	if m.SentAt == nil {
		m.SentAt = timestamppb.Now()
	}
	if m.Status == chatpb.MessageStatus_UNKNOWN {
		m.Status = chatpb.MessageStatus_SENT
	}

	_, err := s.db.Exec(ctx, `
		INSERT INTO chat_messages
		    (id, room_id, sender_id, text, sent_at, status)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, m.Id, m.RoomId, m.SenderId, m.Text,
		m.SentAt.AsTime(), int16(m.Status))
	if err != nil {
		return nil, err
	}
	return &chatpb.Empty{}, nil
}

func (s *server) UpdateStatus(ctx context.Context,
	req *chatpb.UpdateStatusRequest) (*chatpb.Empty, error) {

	_, err := s.db.Exec(ctx, `
		UPDATE chat_messages
		SET status = $1
		WHERE id = $2
	`, int16(req.Status), req.Id)
	if err != nil {
		return nil, err
	}
	return &chatpb.Empty{}, nil
}

const (
	user = "user"
	chat = "chat"
	task = "task"
)

var acl = map[string][]string{
	// UserService methods
	"/user.UserService/AddUser":                {user},
	"/user.UserService/GetUserLinks":           {user},
	"/user.UserService/CheckCredentials":       {user},
	"/user.UserService/GetUserByID":            {user},
	"/user.UserService/UserExists":             {user},
	"/user.UserService/UpdateUserProfile":      {user},
	"/user.UserService/UpdateRating":           {user},
	"/user.UserService/AcceptRequest":          {user},
	"/user.UserService/HasTeacher":             {user},
	"/user.UserService/AddRequestLink":         {user},
	"/user.UserService/DenyRequest":            {user},
	"/user.UserService/GetAvailableTeachers":   {user},
	"/user.UserService/GetRating":              {user},
	"/user.UserService/GetRequests":            {user},
	"/user.UserService/GetStudentTeacherLinks": {user},
	"/user.UserService/GetStudentsByTeacher":   {user},
	"/user.UserService/GetTeachersByStudent":   {user},
	"/user.UserService/GetUsersByIDs":          {user},

	// TaskService methods
	"/taskpb.TaskService/CreateTask":       {task},
	"/taskpb.TaskService/GetTask":          {task},
	"/taskpb.TaskService/GetSolution":      {task},
	"/taskpb.TaskService/LinkFileTask":     {task},
	"/taskpb.TaskService/LinkFileSolution": {task},
	"/taskpb.TaskService/Grade":            {task},
	"/taskpb.TaskService/Solve":            {task},
	"/taskpb.TaskService/AvgGrade":         {task},
	"/taskpb.TaskService/AllTasks":         {task},

	// ChatService methods
	"/chatpb.ChatService/CreateRoom":   {chat},
	"/chatpb.ChatService/History":      {chat},
	"/chatpb.ChatService/SendMessage":  {chat},
	"/chatpb.ChatService/UpdateStatus": {chat},
}

// UnaryInterceptor — перехватчик запросов
func UnaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	// получаем метаданные
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	authHeader := md.Get("authorization")
	if len(authHeader) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing token")
	}

	token := strings.TrimPrefix(authHeader[0], "Bearer ")
	role, err := getRoleByToken(token)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, "invalid token")
	}

	allowedRoles, ok := acl[info.FullMethod]
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "method not allowed")
	}

	if !contains(allowedRoles, role) {
		return nil, status.Error(codes.PermissionDenied, "access denied")
	}

	// всё ок → вызываем сам метод
	return handler(ctx, req)
}

// простая имитация проверки токена
func getRoleByToken(token string) (string, error) {
	switch token {
	case "chat-token":
		return chat, nil
	case "user-token":
		return user, nil
	case "task-token":
		return task, nil
	default:
		return "", status.Error(codes.Unauthenticated, "unknown token")
	}
}

func contains(list []string, target string) bool {
	for _, item := range list {
		if item == target {
			return true
		}
	}
	return false
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(".env file not found")
	}

	dbHost := os.Getenv("POSTGRES_HOST")
	dbPort := os.Getenv("POSTGRES_PORT")
	dbUser := os.Getenv("POSTGRES_USER")
	dbPass := os.Getenv("POSTGRES_PASS")
	dbName := os.Getenv("POSTGRES_DB")
	serverPort := os.Getenv("SERVER_PORT")

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		dbUser, dbPass, dbHost, dbPort, dbName)

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		log.Fatalf("unable to connect to database: %v\n", err)
	}

	defer func() {
		err := conn.Close(ctx)
		if err != nil {
			log.Fatalf("failed to close connection: %v\n", err)
		}
	}()

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(UnaryInterceptor))
	server := &server{db: conn}
	userpb.RegisterUserServiceServer(grpcServer, server)
	taskpb.RegisterTaskServiceServer(grpcServer, server)
	chatpb.RegisterChatServiceServer(grpcServer, server)

	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", ":"+serverPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("server is running on port %s", serverPort)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
