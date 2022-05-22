package logger

type Severity int

const (
	SeverityPanic Severity = iota
	SeverityFatal
	SeverityError
	SeverityWarning
	SeverityInfo
	SeverityDebug
	SeverityVerbose
)

// Logger is the interface that wraps the basic Log methods.
type Logger interface {
	WithField(key string, value interface{}) Logger

	Log(severity Severity, args ...interface{})
	Logf(severity Severity, format string, args ...interface{})

	Verbose(args ...interface{})
	Verbosef(format string, args ...interface{})

	Debug(args ...interface{})
	Debugf(format string, args ...interface{})

	Info(args ...interface{})
	Infof(format string, args ...interface{})

	Warn(args ...interface{})
	Warnf(format string, args ...interface{})

	Error(args ...interface{})
	Errorf(format string, args ...interface{})

	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})

	Panic(args ...interface{})
	Panicf(format string, args ...interface{})

	SetOutputFile(path string) error
}
