package empty

import (
	"fmt"

	"github.com/johnnyipcom/androidtool/pkg/logger"
)

type emptyLogger struct {
}

var _ logger.Logger = &emptyLogger{}

func New() logger.Logger {
	return &emptyLogger{}
}

func (l emptyLogger) WithField(key string, value interface{}) logger.Logger {
	return l
}

func (l emptyLogger) WithStackParams(enable bool, skip int) logger.Logger {
	return l
}

func (l emptyLogger) Debug(args ...interface{}) {
}

func (l emptyLogger) Debugf(format string, args ...interface{}) {
}

func (l emptyLogger) Info(args ...interface{}) {
}

func (l emptyLogger) Infof(format string, args ...interface{}) {
}

func (l emptyLogger) Warn(args ...interface{}) {
}

func (l emptyLogger) Warnf(format string, args ...interface{}) {
}

func (l emptyLogger) Error(args ...interface{}) {
}

func (l emptyLogger) Errorf(format string, args ...interface{}) {
}

func (l emptyLogger) Fatal(args ...interface{}) {
	panic(args)
}

func (l emptyLogger) Fatalf(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}

func (l emptyLogger) Panic(args ...interface{}) {
	panic(args)
}

func (l emptyLogger) Panicf(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}

func (l emptyLogger) Log(severity logger.Severity, args ...interface{}) {
	if severity == logger.SeverityPanic {
		panic(args)
	}
}

func (l emptyLogger) Logf(severity logger.Severity, format string, args ...interface{}) {
	if severity == logger.SeverityPanic {
		panic(fmt.Sprintf(format, args...))
	}
}

func (l emptyLogger) Verbose(args ...interface{}) {
}

func (l emptyLogger) Verbosef(format string, args ...interface{}) {
}

func (l emptyLogger) SetOutputFile(file string) error {
	return nil
}
