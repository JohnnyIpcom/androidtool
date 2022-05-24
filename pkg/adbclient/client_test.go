package adbclient

import (
	"context"
	"testing"

	"github.com/johnnyipcom/androidtool/pkg/logger/logrus"
)

func TestDeleteFile(t *testing.T) {
	client, err := NewClient(DefaultPort, logrus.New("./test.log"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = client.Start(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer client.Stop()

	device, err := client.GetDevice("R58N819VF0L")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := client.RemoveFile(device, "/sdcard/video.mp4"); err != nil {
		t.Errorf("DeleteFile() error = %v", err)
	}
}
