package log

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var logger *zap.Logger

type Config struct {
	Level      string
	SendToFile bool
	Filename   string
	MaxSize    int // megabytes
	MaxAge     int // days
	MaxBackups int
}

func Init(cfg *Config) {
	var l = new(zapcore.Level)
	err := l.UnmarshalText([]byte(cfg.Level))
	if err != nil {
		panic(err)
	}

	consoleSyncer := zapcore.AddSync(os.Stdout)
	consoleEncoder := getConsoleEncoder()
	consoleCore := zapcore.NewCore(consoleEncoder, consoleSyncer, l)

	var opts []zap.Option
	opts = append(opts, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel), zap.Development())

	core := consoleCore
	if cfg.SendToFile {
		fileSyncer := getLogWriter(cfg.Filename, cfg.MaxSize, cfg.MaxBackups, cfg.MaxAge)
		fileEncoder := getJSONEncoder()
		fileCore := zapcore.NewCore(fileEncoder, fileSyncer, l)

		core = zapcore.NewTee(consoleCore, fileCore)
	}

	logger = zap.New(core, opts...)
}

func getJSONEncoder() zapcore.Encoder {
	return getEncoder(true)
}

func getConsoleEncoder() zapcore.Encoder {
	return getEncoder(false)
}

func getEncoder(jsonFormat bool) zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	if jsonFormat {
		return zapcore.NewJSONEncoder(encoderConfig)
	}

	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getLogWriter(filename string, maxSize, maxBackup, maxAge int) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSize,
		MaxBackups: maxBackup,
		MaxAge:     maxAge,
	}

	return zapcore.AddSync(lumberJackLogger)
}

func Logger() *zap.Logger {
	return getLogger()
}

func SugaredLogger() *zap.SugaredLogger {
	return getLogger().Sugar()
}

func getLogger() *zap.Logger {
	if logger == nil {
		panic("Logger is not initialized yet!")
	}
	return logger
}

func getSugaredLogger() *zap.SugaredLogger {
	return getLogger().Sugar()
}

func NewFileLogger(path string) *zap.Logger {
	fileSyncer := getLogWriter(path, 0, 0, 0)
	fileEncoder := getJSONEncoder()
	fileCore := zapcore.NewCore(fileEncoder, fileSyncer, zap.DebugLevel)
	return zap.New(fileCore)
}

func With(fields ...zap.Field) *zap.Logger {
	return getLogger().With(fields...)
}

func Info(msg string) {
	getSugaredLogger().Info(msg)
}

func Infof(format string, a ...interface{}) {
	getSugaredLogger().Infof(format, a...)
}

func Warning(msg string) {
	getSugaredLogger().Warn(msg)
}

func Warningf(format string, a ...interface{}) {
	getSugaredLogger().Warnf(format, a...)
}

func Error(msg string) {
	getSugaredLogger().Error(msg)
}

func Errorf(format string, a ...interface{}) {
	getSugaredLogger().Errorf(format, a...)
}

func Fatal(msg string) {
	getSugaredLogger().Fatal(msg)
}

func Fatalf(format string, a ...interface{}) {
	getSugaredLogger().Fatalf(format, a...)
}
