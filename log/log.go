// Package log provides a leveled, rotated, fast, structured logger.
// This package APIs Print and Fatal are compatible the standard library log.
package log

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// GlobalLogger will be inited with config from env vars.
// All funcs in this package use GlobalLogger
var GlobalLogger *zap.SugaredLogger = NewLogger(NewConfigFromEnv())

// Config _
type Config struct {
	// default log level is debug
	IsLogLevelInfo bool
	// default log to stdout
	LogFilePath string
	// whether to log simultaneously to both stdout and file
	IsNotLogBoth bool
	// whether to rotate log file at midnight
	IsNotLogRotate bool
	// default 24 hours (rotate at midnight)
	RotateInterval time.Duration
}

// NewConfigFromEnv reads env vars to return a Config
func NewConfigFromEnv() Config {
	var c Config
	c.IsLogLevelInfo, _ = strconv.ParseBool(os.Getenv("LOG_LEVEL_INFO"))
	c.LogFilePath = os.Getenv("LOG_FILE_PATH")
	c.IsNotLogBoth, _ = strconv.ParseBool(os.Getenv("LOG_NOT_STDOUT"))
	c.IsNotLogRotate, _ = strconv.ParseBool(os.Getenv("LOG_NOT_ROTATE"))
	c.RotateInterval = 24 * time.Hour
	return c
}

// NewLogger returns a inited Logger
func NewLogger(conf Config) *zap.SugaredLogger {
	encoderConf := zap.NewProductionEncoderConfig()
	encoderConf.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewConsoleEncoder(encoderConf)

	var writers []zapcore.WriteSyncer
	stdWriter, _, _ := zap.Open("stdout")
	if conf.LogFilePath == "" {
		writers = []zapcore.WriteSyncer{stdWriter}
	} else {
		var fileWriter zapcore.WriteSyncer
		if conf.IsNotLogRotate {
			fileWriter, _, _ = zap.Open(conf.LogFilePath)
		} else {
			fileWriter = zapcore.AddSync(newTimedRotatingWriter(
				&lumberjack.Logger{Filename: conf.LogFilePath},
				conf.RotateInterval,
			))
		}
		if conf.IsNotLogBoth {
			writers = []zapcore.WriteSyncer{fileWriter}
		} else {
			writers = []zapcore.WriteSyncer{stdWriter, fileWriter}
		}
	}
	combinedWriter := zap.CombineWriteSyncers(writers...)

	logLevel := zap.DebugLevel
	if conf.IsLogLevelInfo {
		logLevel = zap.InfoLevel
	}
	core := zapcore.NewCore(encoder, combinedWriter, logLevel)
	zl := zap.New(core, zap.AddCaller())
	zl = zl.WithOptions(zap.AddCallerSkip(1))
	logger := zl.Sugar()
	return logger
}

type timedRotatingWriter struct {
	*lumberjack.Logger
	interval    time.Duration
	mutex       sync.RWMutex
	lastRotated time.Time
}

func newTimedRotatingWriter(base *lumberjack.Logger, interval time.Duration) *timedRotatingWriter {
	w := &timedRotatingWriter{Logger: base, interval: interval}
	w.mutex.Lock()
	w.Logger.Rotate()
	w.lastRotated = time.Now().Truncate(interval)
	w.mutex.Unlock()
	return w
}

func (w *timedRotatingWriter) rotateIfNeeded() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if time.Now().Sub(w.lastRotated) < w.interval {
		return nil
	}
	w.lastRotated = time.Now().Truncate(w.interval)
	fmt.Printf("%v about to rotate log file\n", w.lastRotated)
	err := w.Logger.Rotate()
	return err
}

func (w *timedRotatingWriter) Write(p []byte) (int, error) {
	err := w.rotateIfNeeded()
	if err != nil {
		return 0, err
	}
	// ensure no goroutine write log while rotating
	w.mutex.RLock()
	n, err := w.Logger.Write(p)
	w.mutex.RUnlock()
	return n, err
}

func Fatal(args ...interface{}) {
	GlobalLogger.Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	GlobalLogger.Fatalf(format, args...)
}

func Info(args ...interface{}) {
	GlobalLogger.Info(args...)
}

func Infof(format string, args ...interface{}) {
	GlobalLogger.Infof(format, args...)
}

func Debug(args ...interface{}) {
	GlobalLogger.Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	GlobalLogger.Debugf(format, args...)
}

func Print(args ...interface{}) {
	GlobalLogger.Info(args...)
}

func Println(args ...interface{}) {
	GlobalLogger.Info(args...)
}

func Printf(format string, args ...interface{}) {
	GlobalLogger.Infof(format, args...)
}

func Condf(cond bool, format string, args ...interface{}) {
	if cond {
		GlobalLogger.Infof(format, args...)
	}
}

// StdLogger is compatible with the standard library logger,
// This logger call the GlobalLogger funcs
type StdLogger struct{}

func padArgs(args []interface{}) []interface{} {
	if len(args) <= 1 {
		return args
	}
	newArgs := make([]interface{}, 2*len(args)-1)
	for i, e := range args {
		newArgs[2*i] = e
		if i != len(args)-1 {
			newArgs[2*i+1] = " "
		}
	}
	return newArgs
}

func (l StdLogger) Print(args ...interface{}) {
	GlobalLogger.Info(padArgs(args)...)
}

func (l StdLogger) Println(args ...interface{}) {
	GlobalLogger.Info(padArgs(args)...)
}

func (l StdLogger) Printf(format string, args ...interface{}) {
	GlobalLogger.Infof(format, args...)
}

func (l *StdLogger) Fatal(v ...interface{}) {
	GlobalLogger.Fatal(padArgs(v)...)
}

func (l *StdLogger) Fatalln(v ...interface{}) {
	GlobalLogger.Fatal(padArgs(v)...)
}

func (l *StdLogger) Fatalf(format string, v ...interface{}) {
	GlobalLogger.Fatalf(format, v...)
}

func (l *StdLogger) Panic(v ...interface{}) {
	GlobalLogger.Info(v...)
	panic(920911)
}

func (l *StdLogger) Panicf(format string, v ...interface{}) {
	GlobalLogger.Infof(format, v...)
	panic(920911)
}

func (l *StdLogger) Panicln(v ...interface{}) {
	GlobalLogger.Info(v...)
	panic(920911)
}
