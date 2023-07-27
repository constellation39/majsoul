package logger

import (
	"log"
	"runtime"
	"sync/atomic"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	development   bool = false
	logger        *zap.Logger
	atomicLevel   = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	currentConfig zap.Config
	closeFlag     int32
)

func Init() func() {
	SetOutput("stdout")
	SetErrorOutput("stderr")
	EnableDevelopment()
	return Sync
}

func init() {
	currentConfig = zap.Config{
		Level:             atomicLevel,
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: false,
		Sampling: &zap.SamplingConfig{
			Initial:    0,
			Thereafter: 1,
		},
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{},
		ErrorOutputPaths: []string{},
		InitialFields:    nil,
	}
	updateLoggerCore()
}

// EnableDevelopment 启动开发模式
func EnableDevelopment() {
	development = true
	updateLoggerCore()
}

// EnableProduction 启动生产模式
func EnableProduction() {
	development = false
	updateLoggerCore()
}

// SetOutput 设置日志输出到控制台
func SetOutput(output ...string) {
	currentConfig.OutputPaths = output
	updateLoggerCore()
}

// SetErrorOutput 设置日志输出到文件
func SetErrorOutput(errorOutput ...string) {
	currentConfig.ErrorOutputPaths = errorOutput
	updateLoggerCore()
}

func updateLoggerCore() {
	if development {
		atomicLevel.SetLevel(zap.DebugLevel)
		developmentEncoderConfig := zap.NewDevelopmentEncoderConfig()
		developmentEncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.99")
		developmentEncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		currentConfig.EncoderConfig = developmentEncoderConfig
	} else {
		atomicLevel.SetLevel(zap.ErrorLevel)
		productionEncoderConfig := zap.NewProductionEncoderConfig()
		productionEncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.99")
		productionEncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		currentConfig.EncoderConfig = productionEncoderConfig
	}
	if runtime.GOOS == "windows" {
		currentConfig.Encoding = "console"
	}
	var err error
	logger, err = currentConfig.Build(zap.AddStacktrace(zap.ErrorLevel), zap.AddCallerSkip(1))
	if err != nil {
		log.Fatalf("Failed: initLogger error %+v", err)
	}
}

func Sync() {
	if atomic.LoadInt32(&closeFlag) == 1 {
		atomic.CompareAndSwapInt32(&closeFlag, 0, 1)
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
