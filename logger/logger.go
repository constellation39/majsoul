package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"runtime"
	"sync/atomic"
)

var logger = struct {
	*zap.Logger
	closeFlag int32
}{}

func init() {
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
	logger.Logger, err = config.Build(zap.AddStacktrace(zap.ErrorLevel), zap.AddCallerSkip(1))
	if err != nil {
		log.Fatalf("Failed: initLogger error %+v", err)
	}
}

func EnableDevelopment() {
	initLogger(true)
}

func EnableProduction() {
	initLogger(false)
}

func Sync() {
	if atomic.LoadInt32(&logger.closeFlag) == 1 {
		atomic.CompareAndSwapInt32(&logger.closeFlag, 0, 1)
		err := logger.Sync()
		if err != nil {
			log.Printf("%+v", err)
			return
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
