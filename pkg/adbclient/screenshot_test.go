package adbclient

import (
	"context"
	"testing"

	"github.com/johnnyipcom/androidtool/pkg/logger/logrus"
)

func TestScreenshot(t *testing.T) {
	client, err := NewClient(DefaultPort, logrus.New("./test.log"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = client.Start(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer client.Stop()

	device, err := client.GetDevice("133fc498")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	screenshotPath := client.GetScreenshotPath()
	if err := client.Screenshot(device, screenshotPath, WithScreenshotAsPng()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = client.Download(context.Background(), device, screenshotPath, "./test.png")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
