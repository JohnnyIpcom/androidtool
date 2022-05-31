package storage

import (
	"encoding/json"
	"time"

	"github.com/johnnyipcom/androidtool/pkg/adbclient"
	"github.com/johnnyipcom/androidtool/pkg/logger"
	"go.etcd.io/bbolt"
)

const (
	// DefaultStoragePath is the default storage path.
	DefaultStoragePath = "./storage.db"

	// DeviceBucket is the name of the bucket for devices.
	DeviceBucket = "devices"
)

// Storage is the storage for androidtool.
type Storage struct {
	db  *bbolt.DB
	log logger.Logger
}

// NewStorage creates a new storage on the given path.
func NewStorage(path string, log logger.Logger) (*Storage, error) {
	innerLog := log.WithField("package", "storage")

	innerLog.Infof("Opening database at %s", path)
	db, err := bbolt.Open(path, 0600, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	if err := db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(DeviceBucket))
		return err
	}); err != nil {
		return nil, err
	}

	return &Storage{
		db:  db,
		log: innerLog,
	}, nil
}

// Close closes the storage.
func (s *Storage) Close() error {
	return s.db.Close()
}

// SaveDevice creates a new device.
func (s *Storage) SaveDevice(device *adbclient.Device) error {
	s.log.Infof("New device: %s", device.Serial)

	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(DeviceBucket))
		if b == nil {
			return nil
		}

		data, err := json.Marshal(device)
		if err != nil {
			return err
		}

		return b.Put([]byte(device.Serial), data)
	})
}

// GetDevice returns the device with the given serial.
func (s *Storage) GetDevice(serial string) (*adbclient.Device, error) {
	s.log.Infof("Getting device: %s", serial)

	var device *adbclient.Device
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(DeviceBucket))
		if b == nil {
			return nil
		}

		data := b.Get([]byte(serial))
		if data == nil {
			return nil
		}

		device = &adbclient.Device{State: adbclient.StateInvalid}
		return json.Unmarshal(data, &device)
	})

	return device, err
}

// GetDevices returns all devices.
func (s *Storage) GetDevices() ([]*adbclient.Device, error) {
	s.log.Info("Getting devices")

	var devices []*adbclient.Device
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(DeviceBucket))
		if b == nil {
			return nil
		}

		return b.ForEach(func(k, v []byte) error {
			device := &adbclient.Device{State: adbclient.StateInvalid}
			if err := json.Unmarshal(v, &device); err != nil {
				return err
			}

			devices = append(devices, device)
			return nil
		})
	})

	return devices, err
}

// DeleteDevice deletes the device with the given serial.
func (s *Storage) DeleteDevice(serial string) error {
	s.log.Infof("Deleting device: %s", serial)

	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(DeviceBucket))
		if b == nil {
			return nil
		}

		return b.Delete([]byte(serial))
	})
}
