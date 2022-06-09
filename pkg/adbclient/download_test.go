package adbclient

import (
	"context"
	"testing"

	"github.com/johnnyipcom/androidtool/pkg/logger/logrus"
)

func TestDownloadFile(t *testing.T) {
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

	err = client.DownloadFile(context.Background(), device, "/sdcard/Kik/4e17b973-b4f2-4f3a-860f-bfbac196aadc.jpg", "./test.jpg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
