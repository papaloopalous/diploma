package repo

import (
	"api/internal/proto/chatpb"
	"time"

	"github.com/google/uuid"
)

// UsersList содержит информацию о пользователе
type UsersList struct {
	ID        uuid.UUID `json:"id"`                  // Уникальный идентификатор пользователя
	Fio       string    `json:"fio"`                 // ФИО пользователя
	Age       uint8     `json:"age"`                 // Возраст пользователя
	Specialty string    `json:"specialty,omitempty"` // Специализация преподавателя
	Price     int       `json:"price,omitempty"`     // Стоимость занятия
	Rating    float32   `json:"rating"`              // Рейтинг преподавателя
}

const (
	authorization = "authorization"
	bearer        = "Bearer "
)

// UserRepo определяет методы для работы с пользователями в системе
type UserRepo interface {
	// FindUser находит пользователя по ID
	FindUser(userID uuid.UUID) (user UsersList, err error)

	// CheckPass проверяет учетные данные пользователя
	CheckPass(username string, pass string) (userID uuid.UUID, role string, err error)

	// CreateAccount создает новую учетную запись
	CreateAccount(username string, pass string, role string) (userID uuid.UUID, err error)

	// OutAscendingBySpecialty возвращает отсортированный по возрастанию список преподавателей
	OutAscendingBySpecialty(orderField string, specialty string, userID uuid.UUID) (users []UsersList, err error)

	// OutDescendingBySpecialty возвращает отсортированный по убыванию список преподавателей
	OutDescendingBySpecialty(orderField string, specialty string, userID uuid.UUID) (users []UsersList, err error)

	// HasThatTeacher проверяет связь студента с преподавателем
	HasThatTeacher(studentID uuid.UUID, teacherID uuid.UUID) (bool, error)

	// AddRating добавляет оценку преподавателю
	AddRating(userID uuid.UUID, rating float32) error

	// StudentsByTeacher возвращает список студентов преподавателя
	StudentsByTeacher(teacherID uuid.UUID) (users []UsersList, err error)

	// EditGrade обновляет среднюю оценку студента
	EditGrade(studentID uuid.UUID, grade float32) error

	// FillProfile обновляет профиль пользователя
	FillProfile(userID uuid.UUID, userData UsersList) error

	// TeachersByStudent возвращает список преподавателей студента
	TeachersByStudent(studentID uuid.UUID) (teachers []UsersList, err error)

	// AddRequest создает запрос на обучение
	AddRequest(studentID uuid.UUID, teacherID uuid.UUID) error

	// ShowRequests возвращает список запросов на обучение
	ShowRequests(userID uuid.UUID) (users []UsersList, err error)

	// Accept подтверждает запрос на обучение
	Accept(teacherID uuid.UUID, studentID uuid.UUID) error

	// Deny отклоняет запрос на обучение
	Deny(teacherID uuid.UUID, studentID uuid.UUID) error
}

// taskList содержит информацию о задании
type taskList struct {
	ID      uuid.UUID `json:"taskID"`          // Уникальный идентификатор задания
	Name    string    `json:"taskName"`        // Название задания
	Grade   uint8     `json:"grade,omitempty"` // Оценка за задание
	Status  string    `json:"status"`          // Статус задания
	Student string    `json:"student"`         // ФИО студента
	Teacher string    `json:"teacher"`         // ФИО преподавателя
}

// TaskRepo определяет методы для работы с заданиями
type TaskRepo interface {
	// CreateTask создает новое задание
	CreateTask(teacher uuid.UUID, student uuid.UUID, name string, studentFIO string, teacherFIO string) (uuid.UUID, error)

	// GetTask получает файл задания
	GetTask(taskID uuid.UUID) (fileName string, fileData []byte, err error)

	// GetSolution получает файл решения
	GetSolution(taskID uuid.UUID) (fileName string, fileData []byte, err error)

	// LinkFileTask прикрепляет файл к заданию
	LinkFileTask(taskID uuid.UUID, fileName string, fileData []byte) error

	// LinkFileSolution прикрепляет файл решения
	LinkFileSolution(taskID uuid.UUID, fileName string, fileData []byte) error

	// Grade выставляет оценку за задание
	Grade(taskID uuid.UUID, grade uint8) (studentID uuid.UUID, err error)

	// Solve отмечает задание как решенное
	Solve(taskID uuid.UUID) error

	// AvgGrade считает среднюю оценку студента
	AvgGrade(studentID uuid.UUID) (grade float32, err error)

	// AllTasks возвращает все задания пользователя
	AllTasks(userID uuid.UUID) (tasks []taskList)
}

// SessionRepo определяет методы для работы с сессиями
type SessionRepo interface {
	// GetSession получает информацию о сессии
	GetSession(sessionID uuid.UUID) (userID uuid.UUID, role string, err error)

	// SetSession создает новую сессию
	SetSession(sessionID uuid.UUID, userID uuid.UUID, role string, sessionLifetime time.Duration) error

	// DeleteSession удаляет сессию
	DeleteSession(sessionID uuid.UUID) (userID uuid.UUID, err error)
}

// ChatMessage содержит информацию о сообщении в чате
type ChatMessage struct {
	ID       uuid.UUID            // Уникальный идентификатор сообщения
	RoomID   string               // Идентификатор комнаты
	SenderID uuid.UUID            // Идентификатор отправителя
	Text     string               // Текст сообщения
	SentAt   time.Time            // Время отправки
	Status   chatpb.MessageStatus // Статус сообщения
}

// ChatRepo определяет методы для работы с чатом
type ChatRepo interface {
	// CreateRoom создает новую комнату чата
	CreateRoom(user1, user2 uuid.UUID) (roomID string, existed bool, err error)

	// History возвращает историю сообщений
	History(roomID string) ([]ChatMessage, error)

	// SaveMessage сохраняет новое сообщение
	SaveMessage(msg ChatMessage) error

	// UpdateStatus обновляет статус сообщения
	UpdateStatus(msgID uuid.UUID, status chatpb.MessageStatus) error
}
