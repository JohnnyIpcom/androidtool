package storage_test

import (
	"os"
	"testing"

	"github.com/johnnyipcom/androidtool/internal/storage"
	"github.com/johnnyipcom/androidtool/pkg/adbclient"
	"github.com/johnnyipcom/androidtool/pkg/logger/empty"
)

func TestStorage(t *testing.T) {
	storage, err := storage.NewStorage("/temp.db", empty.New())
	if err != nil {
		t.Error(err)
	}

	defer func() {
		storage.Close()

		//remove the file
		os.Remove("/temp.db")
	}()

	oldDevices := []*adbclient.Device{
		{
			Serial: "123456789",
			State:  adbclient.StateOnline,
			Model:  "Nexus 5",
			Display: adbclient.DisplayParams{
				Width:   320,
				Height:  480,
				Density: 160,
			},
		},
		{
			Serial: "987654321",
			State:  adbclient.StateOffline,
			Model:  "Samsung Galaxy S8",
			Display: adbclient.DisplayParams{
				Width:   320,
				Height:  480,
				Density: 160,
			},
		},
	}

	for _, device := range oldDevices {
		if err := storage.SaveDevice(device); err != nil {
			t.Error(err)
		}
	}

	newDevices, err := storage.GetDevices()
	if err != nil {
		t.Error(err)
	}

	if len(newDevices) != 2 {
		t.Error("Expected 2 devices")
	}

	if newDevices[0].Serial != "123456789" {
		t.Error("Expected serial 123456789")
	}

	if newDevices[1].Serial != "987654321" {
		t.Error("Expected serial 987654321")
	}

	if newDevices[0].Model != "Nexus 5" {
		t.Error("Expected model Nexus 5")
	}

	if newDevices[1].Model != "Samsung Galaxy S8" {
		t.Error("Expected model Samsung Galaxy S8")
	}

	if newDevices[0].Display.Width != 320 {
		t.Error("Expected display width 320")
	}

	if newDevices[1].Display.Width != 320 {
		t.Error("Expected display width 320")
	}

	device0, err := storage.GetDevice("123456789")
	if err != nil {
		t.Error(err)
	}

	if device0.Serial != "123456789" {
		t.Error("Expected serial 123456789")
	}

	if device0.Model != "Nexus 5" {
		t.Error("Expected model Nexus 5")
	}

	if device0.Display.Width != 320 {
		t.Error("Expected display width 320")
	}

	for _, device := range oldDevices {
		if err := storage.DeleteDevice(device.Serial); err != nil {
			t.Error(err)
		}
	}
}
