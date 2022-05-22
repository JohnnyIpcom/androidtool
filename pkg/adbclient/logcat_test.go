package adbclient

import (
	"testing"
	"time"
)

func TestParseLogcatMessage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected LogcatMessage
	}{
		{
			name:  "test1",
			input: `05-18 12:01:09.830  5233  7998 D MARsPolicyManager: onPackageResumedFG pkgName = com.sec.android.app.clockpackage, userId = 0`,
			expected: LogcatMessage{
				Timestamp: time.Date(time.Now().Year(), time.May, 18, 12, 1, 9, 830*1e6, time.UTC),
				Priority:  Debug,
				Tag:       "MARsPolicyManager",
				ProcessID: 5233,
				ThreadID:  7998,
				Message:   "onPackageResumedFG pkgName = com.sec.android.app.clockpackage, userId = 0",
			},
		},
		{
			name:  "test2",
			input: `05-18 12:01:09.864  5233  5278 I GameSDK@LifeCycle: noteResumeComponent(): package name  : com.sec.android.app.clockpackage`,
			expected: LogcatMessage{
				Timestamp: time.Date(time.Now().Year(), time.May, 18, 12, 1, 9, 864*1e6, time.UTC),
				Priority:  Info,
				Tag:       "GameSDK@LifeCycle",
				ProcessID: 5233,
				ThreadID:  5278,
				Message:   "noteResumeComponent(): package name  : com.sec.android.app.clockpackage",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := ParseLogcatMessage(test.input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if actual != test.expected {
				t.Errorf("expected: %v, actual: %v", test.expected, actual)
			}
		})
	}
}
