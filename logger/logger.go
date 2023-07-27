package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"runtime"
	"sync/atomic"
)

// Global variables for configuration and state maintenance
var (
	development   bool                                       = false // Whether it is in development mode
	logger        *zap.Logger                                        // Log object from zap library
	atomicLevel   = zap.NewAtomicLevelAt(zapcore.DebugLevel)         // Log level
	currentConfig zap.Config                                         // Configuration object from zap library
	closeFlag     int32                                              // Close flag
)

// Init initializes the logger and returns a Sync function for synchronizing log records
func Init() func() {
	SetOutput("stdout")      // Set log output to console
	SetErrorOutput("stderr") // Set error log output to console error output
	EnableDevelopment()      // Enable development mode
	return Sync              // Return Sync function for synchronizing log records
}

// init updates the core configuration of logger when the package is initialized
func init() {
	currentConfig = zap.Config{
		Level:             atomicLevel,
		Development:       development,
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

// EnableDevelopment enables development mode
func EnableDevelopment() {
	development = true
	updateLoggerCore()
}

// EnableProduction enables production mode
func EnableProduction() {
	development = false
	updateLoggerCore()
}

// SetOutput sets the log output to the specified path
func SetOutput(output ...string) {
	currentConfig.OutputPaths = output
	updateLoggerCore()
}

// SetErrorOutput sets the error log output to the specified path
func SetErrorOutput(errorOutput ...string) {
	currentConfig.ErrorOutputPaths = errorOutput
	updateLoggerCore()
}

// updateLoggerCore updates the core configuration of the zap library, including log level, time format, caller information, etc.
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

// Sync synchronizes log records to ensure that all logs are output
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

// Debug records Debug level logs
func Debug(msg string, fields ...zap.Field) {
	logger.Debug(msg, fields...)
}

// Info records Info level logs
func Info(msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}

// Warn records Warn level logs
func Warn(msg string, fields ...zap.Field) {
	logger.Warn(msg, fields...)
}

// Error records Error level logs
func Error(msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
}

// DPanic records DPanic level logs
func DPanic(msg string, fields ...zap.Field) {
	logger.DPanic(msg, fields...)
}

// Panic records Panic level logs
func Panic(msg string, fields ...zap.Field) {
	logger.Panic(msg, fields...)
}

// Fatal records Fatal level logs
func Fatal(msg string, fields ...zap.Field) {
	logger.Fatal(msg, fields...)
}
