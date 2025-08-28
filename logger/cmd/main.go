package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	"logger/logservice"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

// logger глобальный экземпляр логгера
var logger *zap.Logger

// LogServer реализует gRPC сервер для логирования
type LogServer struct {
	logservice.UnimplementedLogServiceServer
}

// WriteLog обрабатывает запросы на логирование
// Записывает логи в консоль и файл с указанным уровнем важности
func (s *LogServer) WriteLog(ctx context.Context, req *logservice.LogRequest) (*logservice.LogResponse, error) {
	sugaredLogger := zap.SugaredLogger(*logger.Sugar())

	// Добавляем название сервиса в метаданные лога
	logEntry := sugaredLogger.With(
		"service", req.Service,
	)

	// Добавляем дополнительные метаданные
	for k, v := range req.Metadata {
		logEntry = logEntry.With(k, v)
	}

	// Логируем сообщение с соответствующим уровнем
	switch req.Level {
	case "INFO":
		logEntry.Info(req.Message)
	case "ERROR":
		logEntry.Error(req.Message)
	case "DEBUG":
		logEntry.Debug(req.Message)
	default:
		logEntry.Warn("unknown log level")
	}

	return &logservice.LogResponse{Success: true}, nil
}

func main() {
	// Создаем конфигурацию для многоуровневого логирования
	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Создаем директорию для логов если её нет
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("failed to create log directory: %v", err)
	}

	// Открываем файл для логов
	logFile, err := os.OpenFile(
		filepath.Join(logDir, "service.log"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}

	// Создаем ядро логгера для консоли
	consoleCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.AddSync(os.Stdout),
		zap.InfoLevel,
	)

	// Создаем ядро логгера для файла
	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.AddSync(logFile),
		zap.InfoLevel,
	)

	// Объединяем ядра
	core := zapcore.NewTee(consoleCore, fileCore)

	// Создаем логгер
	logger = zap.New(core)
	defer func() {
		err := logger.Sync()
		if err != nil {
			log.Printf("failed to sync logger: %v", err)
		}
	}()

	// Загружаем переменные окружения
	err = godotenv.Load()
	if err != nil {
		log.Fatal(".env was not found")
	}

	// Получаем порт из переменных окружения
	grpcPort := os.Getenv("LOGGER_PORT")

	// Запускаем gRPC сервер
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("error starting a server: %v", err)
	}
	grpcServer := grpc.NewServer()
	logservice.RegisterLogServiceServer(grpcServer, &LogServer{})

	fmt.Println("LogService is running on " + grpcPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("grpc error: %v", err)
	}
}
