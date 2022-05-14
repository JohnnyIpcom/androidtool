package logger

// Logger is the interface that wraps the basic Log methods.
type Logger interface {
	WithField(key string, value interface{}) Logger

	Info(args ...interface{})
	Infof(format string, args ...interface{})

	Debug(args ...interface{})
	Debugf(format string, args ...interface{})

	Warn(args ...interface{})
	Warnf(format string, args ...interface{})

	Error(args ...interface{})
	Errorf(format string, args ...interface{})

	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})

	SetOutputFile(path string) error
}
