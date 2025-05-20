package logger

import (
	"log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

func Init() {
	level := zapcore.InfoLevel

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	consoleEncoder := zapcore.NewConsoleEncoder(encoderCfg)
	fileEncoder := zapcore.NewJSONEncoder(encoderCfg)

	logFile, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Println("cannot open log file: " + err.Error())
	}

	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), level)
	fileCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(logFile), level)

	core := zapcore.NewTee(consoleCore, fileCore)

	Log = zap.New(core, zap.AddCaller())
}
