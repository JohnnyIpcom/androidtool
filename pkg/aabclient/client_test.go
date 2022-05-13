package aabclient

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
)

func TestBuildAPK(t *testing.T) {
	c, err := NewClient("1.10.0", log.New(os.Stdout, "", 0))
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
