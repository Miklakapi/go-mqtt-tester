package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/Miklakapi/go-mqtt-tester/internal/mqtt"
	"github.com/Miklakapi/go-mqtt-tester/internal/watcher"
)

type Controller struct {
	watcher  *watcher.Watcher
	mqtt     *mqtt.MQTTClient
	settings Settings
}

type Settings struct {
	DiscoveryPrefix string
	StatePrefix     string

	DeviceID           string
	DeviceName         string
	DeviceManufacturer string
	DeviceModel        string

	DataDir string
}

type fileData struct {
	fileName string
	sensorID string
	path     string
	data     fileContent
}

type fileContent struct {
	Config map[string]any `json:"config"`
	State  any            `json:"state"`
}

func New(watcher *watcher.Watcher, mqtt *mqtt.MQTTClient, settings Settings) (*Controller, error) {
	if settings.DiscoveryPrefix == "" {
		return nil, ErrMissingDiscoveryPrefix
	}
	if settings.StatePrefix == "" {
		return nil, ErrMissingStatePrefix
	}
	if settings.DeviceID == "" {
		return nil, ErrMissingDeviceID
	}
	if settings.DeviceName == "" {
		return nil, ErrMissingDeviceName
	}
	if settings.DeviceManufacturer == "" {
		return nil, ErrMissingDeviceManufacturer
	}
	if settings.DeviceModel == "" {
		return nil, ErrMissingDeviceModel
	}
	if settings.DataDir == "" {
		return nil, ErrMissingDataDir
	}

	return &Controller{
		watcher:  watcher,
		mqtt:     mqtt,
		settings: settings,
	}, nil
}

func (c *Controller) Run(ctx context.Context) error {
	c.watcher.AddWatch(c.settings.DataDir, syscall.IN_DELETE|syscall.IN_MOVED_FROM|syscall.IN_MOVED_TO|syscall.IN_CLOSE_WRITE)

	for {
		select {
		case event := <-c.watcher.Events:
			if event.Mask&syscall.IN_ISDIR != 0 {
				continue
			}
			fileData, err := c.getFileData(event)
			if err != nil {
				log.Println(err)
				continue
			}
			log.Println(fileData)
			// Publish config
			// c.mqtt.Publish("", 0, false, "")
			// Publish state
			// c.mqtt.Publish("", 0, false, "")
		case err := <-c.watcher.Errors:
			log.Println(err)
		case <-ctx.Done():
			return nil
		}
	}
}

func (c *Controller) getFileData(event watcher.Event) (fileData, error) {
	sensorID := strings.TrimSuffix(event.Name, ".json")
	path := filepath.Join(c.settings.DataDir, event.Name)

	b, err := os.ReadFile(path)
	if err != nil {
		return fileData{}, fmt.Errorf("%w: %s: %w", ErrReadFile, path, err)
	}

	var content fileContent
	if err := json.Unmarshal(b, &content); err != nil {
		return fileData{}, fmt.Errorf("%w: %s: %w", ErrInvalidJSON, path, err)
	}

	return fileData{
		fileName: event.Name,
		sensorID: sensorID,
		path:     path,
		data:     content,
	}, nil
}
