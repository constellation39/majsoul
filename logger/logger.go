package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

var logger *zap.Logger

func init() {
	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	go signalLoop(signalChan)
	initLogger(false)
}

func initLogger(debug bool) {
	developmentEncoderConfig := zap.NewDevelopmentEncoderConfig()
	developmentEncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.99")
	developmentEncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(zap.DebugLevel),
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: false,
		Sampling: &zap.SamplingConfig{
			Initial:    0,
			Thereafter: 1,
		},
		Encoding:         "json",
		EncoderConfig:    developmentEncoderConfig,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		InitialFields:    nil,
	}
	if !debug {
		config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
		productionEncoderConfig := zap.NewProductionEncoderConfig()
		productionEncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.99")
		productionEncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		config.EncoderConfig = productionEncoderConfig

	}
	if runtime.GOOS == "windows" {
		config.Encoding = "console"
	}
	var err error
	logger, err = config.Build(zap.AddStacktrace(zap.ErrorLevel), zap.AddCallerSkip(1))
	if err != nil {
		log.Fatalf("Failed: initLogger error %+v", err)
	}
}

func signalLoop(ch chan os.Signal) {
	for {
		select {
		case sign := <-ch:
			switch sign {
			case syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT:
				if logger != nil {
					logger.Sync()
				}
				return
			}
		}
	}
}

func Debug(msg string, fields ...zap.Field) {
	logger.Debug(msg, fields...)
}
func Info(msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}
func Warn(msg string, fields ...zap.Field) {
	logger.Warn(msg, fields...)
}
func Error(msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
}
func DPanic(msg string, fields ...zap.Field) {
	logger.DPanic(msg, fields...)
}
func Panic(msg string, fields ...zap.Field) {
	logger.Panic(msg, fields...)
}
func Fatal(msg string, fields ...zap.Field) {
	logger.Fatal(msg, fields...)
}
