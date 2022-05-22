package adbclient

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"time"

	"github.com/johnnyipcom/androidtool/pkg/logger"
	"github.com/zach-klippenstein/goadb/wire"
)

var (
	ErrInvalidLogcatMessage  = fmt.Errorf("invalid logcat message")
	ErrInvalidLogcatPriority = fmt.Errorf("invalid logcat priority")
)

type LogcatPriority int

const (
	// Verbose represents the 'V' logcat message priority.
	Verbose LogcatPriority = iota
	// Debug represents the 'D' logcat message priority.
	Debug
	// Info represents the 'I' logcat message priority.
	Info
	// Warning represents the 'W' logcat message priority.
	Warning
	// Error represents the 'E' logcat message priority.
	Error
	// Fatal represents the 'F' logcat message priority.
	Fatal
)

func (l LogcatPriority) Severity() logger.Severity {
	switch l {
	case Verbose:
		return logger.SeverityVerbose
	case Debug:
		return logger.SeverityDebug
	case Info:
		return logger.SeverityInfo
	case Warning:
		return logger.SeverityWarning
	case Error:
		return logger.SeverityError
	case Fatal:
		return logger.SeverityFatal
	default:
		return logger.SeverityPanic
	}
}

func (l LogcatPriority) String() string {
	switch l {
	case Verbose:
		return "V"
	case Debug:
		return "D"
	case Info:
		return "I"
	case Warning:
		return "W"
	case Error:
		return "E"
	case Fatal:
		return "F"
	default:
		return "S"
	}
}

// LogcatMessage is a representation of a message from adb logcat.
type LogcatMessage struct {
	Timestamp time.Time
	Priority  LogcatPriority
	Tag       string
	ProcessID int
	ThreadID  int
	Message   string
}

func (m LogcatMessage) Log(log logger.Logger) {
	log.Log(m.Severity(), fmt.Sprintf("%s: %s", m.Tag, m.Message))
}

func (m LogcatMessage) Severity() logger.Severity {
	return m.Priority.Severity()
}

func ParseLogcatPriority(s string) (LogcatPriority, error) {
	switch s {
	case "V":
		return Verbose, nil
	case "D":
		return Debug, nil
	case "I":
		return Info, nil
	case "W":
		return Warning, nil
	case "E":
		return Error, nil
	case "F":
		return Fatal, nil
	default:
		return Verbose, ErrInvalidLogcatPriority
	}
}

var logcatMsgRegex = regexp.MustCompile(`\s*([0-9]*)-([0-9]*)\s*([0-9]*):([0-9]*):([0-9]*).([0-9]*)\s*([0-9]*)\s*([0-9]*)\s*([VDIWEF])\s*(.*):\s*(.*)`)

// ParseLogcatMessage parses a logcat message.
func ParseLogcatMessage(msg string) (LogcatMessage, error) {
	parts := logcatMsgRegex.FindStringSubmatch(msg)
	if len(parts) < 12 {
		return LogcatMessage{}, ErrInvalidLogcatMessage
	}
	month, _ := strconv.Atoi(parts[1])
	day, _ := strconv.Atoi(parts[2])
	hour, _ := strconv.Atoi(parts[3])
	minute, _ := strconv.Atoi(parts[4])
	second, _ := strconv.Atoi(parts[5])
	microseconds, _ := strconv.Atoi(parts[6])
	pid, _ := strconv.Atoi(parts[7])
	tid, _ := strconv.Atoi(parts[8])
	priority, _ := ParseLogcatPriority(parts[9])
	tag := parts[10]
	message := parts[11]

	return LogcatMessage{
		Timestamp: time.Date(time.Now().Year(), time.Month(month), day, hour, minute, second, microseconds*1e6, time.Local),
		ProcessID: pid,
		ThreadID:  tid,
		Priority:  priority,
		Tag:       tag,
		Message:   message,
	}, nil
}

// Logcat reads logcat messages from the adb server.
type LogcatWatcher struct {
	reader io.Reader
	conn   *wire.Conn
	log    logger.Logger
}

// Close closes the logcat connection.
func (w *LogcatWatcher) Close() error {
	return w.conn.Close()
}

// C is a channel of LogcatMessages.
func (w *LogcatWatcher) C(ctx context.Context) <-chan LogcatMessage {
	ch := make(chan LogcatMessage)

	go func() {
		for {
			select {
			case <-ctx.Done():
				w.log.Debugf("logcat watcher stopped")
				return

			default:
				reader := bufio.NewReader(w.reader)
				for {
					line, err := reader.ReadString('\n')
					if err != nil {
						w.log.Error(err)
						return
					}

					msg, err := ParseLogcatMessage(line)
					if err != nil {
						w.log.Error(err)
						continue
					}

					ch <- msg
				}
			}
		}
	}()

	return ch
}

// Read implements io.Reader.
func (w *LogcatWatcher) Read(p []byte) (int, error) {
	return w.reader.Read(p)
}
