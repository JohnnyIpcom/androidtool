package logrus

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/johnnyipcom/androidtool/pkg/logger"
	"github.com/sirupsen/logrus"
)

// logrusLogger is a wrapper around logrus.Logger.
type logrusLogger struct {
	log   *logrus.Logger
	entry *logrus.Entry
}

var _ logger.Logger = &logrusLogger{}

func New(path string) logger.Logger {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	log.SetFormatter(&nested.Formatter{
		HideKeys:    true,
		FieldsOrder: []string{"file", "component", "module", "message"},
	})

	l := &logrusLogger{
		log:   log,
		entry: log.WithContext(context.Background()),
	}

	l.SetOutputFile(path)
	return l
}

func (l *logrusLogger) WithField(key string, value interface{}) logger.Logger {
	return &logrusLogger{
		entry: l.entry.WithField(key, value),
	}
}

func (l *logrusLogger) fileInfo(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		file = "<???>"
		line = 1
	} else {
		slash := strings.LastIndex(file, "/")
		if slash >= 0 {
			file = file[slash+1:]
		}
	}

	return fmt.Sprintf("%s:%d", file, line)
}

func (l *logrusLogger) Panic(args ...interface{}) {
	l.entry.Panic(args...)
}

func (l *logrusLogger) Info(args ...interface{}) {
	l.entry.WithField("file", l.fileInfo(2)).Info(args...)
}

func (l *logrusLogger) Infof(format string, args ...interface{}) {
	l.entry.WithField("file", l.fileInfo(2)).Infof(format, args...)
}

func (l *logrusLogger) Debug(args ...interface{}) {
	l.entry.WithField("file", l.fileInfo(2)).Debug(args...)
}

func (l *logrusLogger) Debugf(format string, args ...interface{}) {
	l.entry.WithField("file", l.fileInfo(2)).Debugf(format, args...)
}

func (l *logrusLogger) Warn(args ...interface{}) {
	l.entry.WithField("file", l.fileInfo(2)).Warn(args...)
}

func (l *logrusLogger) Warnf(format string, args ...interface{}) {
	l.entry.WithField("file", l.fileInfo(2)).Warnf(format, args...)
}

func (l *logrusLogger) Error(args ...interface{}) {
	l.entry.WithField("file", l.fileInfo(2)).Error(args...)
}

func (l *logrusLogger) Errorf(format string, args ...interface{}) {
	l.entry.WithField("file", l.fileInfo(2)).Errorf(format, args...)
}

func (l *logrusLogger) Fatal(args ...interface{}) {
	l.entry.WithField("file", l.fileInfo(2)).Fatal(args...)
}

func (l *logrusLogger) Fatalf(format string, args ...interface{}) {
	l.entry.WithField("file", l.fileInfo(2)).Fatalf(format, args...)
}

func (l *logrusLogger) SetOutputFile(path string) error {
	logFile, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		logrus.Errorf("Failed to create log file: %s", err)
		return err
	}

	l.log.SetOutput(io.MultiWriter(logFile, os.Stdout))
	return nil
}
