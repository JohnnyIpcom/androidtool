package aabclient

import (
	"context"
	"fmt"
	"testing"

	"github.com/johnnyipcom/androidtool/pkg/logger/logrus"
)

func TestBuildAPK(t *testing.T) {
	c, err := NewClient("1.10.0", logrus.New("./log.txt"))
	if err != nil {
		t.Fatal(err)
	}

	err = c.Start(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	defer c.Stop()
	keystore := NewDefaultKeystoreConfig("D:\\build.keystore")

	out, err := c.BuildAPKs(context.Background(), "D:\\build.aab", "D:\\build.apks", "R58N819VF0L", keystore)
	if err != nil {
		t.Fatal(fmt.Printf("%s:%s", err.Error(), string(out)))
	}
}
