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

	enableStackTrace bool
	skipStackEntries int
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

		enableStackTrace: true,
		skipStackEntries: 2,
	}

	l.SetOutputFile(path)
	return l
}

func (l *logrusLogger) WithField(key string, value interface{}) logger.Logger {
	return &logrusLogger{
		log:   l.log,
		entry: l.entry.WithField(key, value),

		enableStackTrace: l.enableStackTrace,
		skipStackEntries: l.skipStackEntries,
	}
}

func (l *logrusLogger) WithStackParams(enable bool, skip int) logger.Logger {
	return &logrusLogger{
		log:   l.log,
		entry: l.entry,

		enableStackTrace: enable,
		skipStackEntries: skip,
	}
}

func (l logrusLogger) fileInfo(skip int) string {
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

func (l logrusLogger) toLogrusSeverity(severity logger.Severity) logrus.Level {
	switch severity {
	case logger.SeverityVerbose:
		return logrus.TraceLevel
	case logger.SeverityDebug:
		return logrus.DebugLevel
	case logger.SeverityInfo:
		return logrus.InfoLevel
	case logger.SeverityWarning:
		return logrus.WarnLevel
	case logger.SeverityError:
		return logrus.ErrorLevel
	case logger.SeverityFatal:
		return logrus.FatalLevel
	case logger.SeverityPanic:
		return logrus.PanicLevel
	default:
		return logrus.DebugLevel
	}
}

func (l *logrusLogger) Log(severity logger.Severity, args ...interface{}) {
	l.entry.WithField("file", l.fileInfo(l.skipStackEntries)).Log(l.toLogrusSeverity(severity), args...)
}

func (l *logrusLogger) Logf(severity logger.Severity, format string, args ...interface{}) {
	l.entry.WithField("file", l.fileInfo(l.skipStackEntries)).Logf(l.toLogrusSeverity(severity), format, args...)
}

func (l *logrusLogger) Verbose(args ...interface{}) {
	l.entry.WithField("file", l.fileInfo(l.skipStackEntries)).Trace(args...)
}

func (l *logrusLogger) Verbosef(format string, args ...interface{}) {
	l.entry.WithField("file", l.fileInfo(l.skipStackEntries)).Tracef(format, args...)
}

func (l *logrusLogger) Debug(args ...interface{}) {
	l.entry.WithField("file", l.fileInfo(l.skipStackEntries)).Debug(args...)
}

func (l *logrusLogger) Debugf(format string, args ...interface{}) {
	l.entry.WithField("file", l.fileInfo(l.skipStackEntries)).Debugf(format, args...)
}

func (l *logrusLogger) Info(args ...interface{}) {
	l.entry.WithField("file", l.fileInfo(l.skipStackEntries)).Info(args...)
}

func (l *logrusLogger) Infof(format string, args ...interface{}) {
	l.entry.WithField("file", l.fileInfo(l.skipStackEntries)).Infof(format, args...)
}

func (l *logrusLogger) Warn(args ...interface{}) {
	l.entry.WithField("file", l.fileInfo(l.skipStackEntries)).Warn(args...)
}

func (l *logrusLogger) Warnf(format string, args ...interface{}) {
	l.entry.WithField("file", l.fileInfo(l.skipStackEntries)).Warnf(format, args...)
}

func (l *logrusLogger) Error(args ...interface{}) {
	l.entry.WithField("file", l.fileInfo(l.skipStackEntries)).Error(args...)
}

func (l *logrusLogger) Errorf(format string, args ...interface{}) {
	l.entry.WithField("file", l.fileInfo(l.skipStackEntries)).Errorf(format, args...)
}

func (l *logrusLogger) Fatal(args ...interface{}) {
	l.entry.WithField("file", l.fileInfo(l.skipStackEntries)).Fatal(args...)
}

func (l *logrusLogger) Fatalf(format string, args ...interface{}) {
	l.entry.WithField("file", l.fileInfo(l.skipStackEntries)).Fatalf(format, args...)
}

func (l *logrusLogger) Panic(args ...interface{}) {
	l.entry.WithField("file", l.fileInfo(l.skipStackEntries)).Panic(args...)
}

func (l *logrusLogger) Panicf(format string, args ...interface{}) {
	l.entry.WithField("file", l.fileInfo(l.skipStackEntries)).Panicf(format, args...)
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
