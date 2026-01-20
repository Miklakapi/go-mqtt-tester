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

	device map[string]any
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
	name     string
	sensorID string
	path     string
	data     fileContent
}

type fileContent struct {
	Config    map[string]any `json:"config"`
	Component string         `json:"component"`
	State     any            `json:"state"`
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
		device: map[string]any{
			"identifiers":  []string{settings.DeviceID},
			"name":         settings.DeviceName,
			"manufacturer": settings.DeviceManufacturer,
			"model":        settings.DeviceModel,
		},
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

			if !validateFile(fileData) {
				continue
			}

			configPayload, err := json.Marshal(fileData.data.Config)
			if err != nil {
				log.Println("config marshal error:", err)
				continue
			}

			var statePayload []byte
			if fileData.data.Component == "binary_sensor" {
				s, ok := fileData.data.State.(string)
				if !ok {
					log.Println("binary_sensor state must be string ON/OFF in", fileData.name)
					continue
				}
				statePayload = []byte(s)
			} else {
				var err error
				statePayload, err = json.Marshal(fileData.data.State)
				if err != nil {
					log.Println("state marshal error:", err)
					continue
				}
			}

			c.publishConfig(fileData, configPayload)
			c.publishState(fileData, statePayload)
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

	content.Config["unique_id"] = c.settings.DeviceID + "_" + content.Component + "_" + sensorID
	content.Config["state_topic"] = c.stateTopic(sensorID)
	content.Config["device"] = c.device

	return fileData{
		name:     event.Name,
		sensorID: sensorID,
		path:     path,
		data:     content,
	}, nil
}

func (c *Controller) publishConfig(f fileData, payload []byte) error {
	topic := c.configTopic(f.data.Component, f.sensorID)
	return c.mqtt.Publish(topic, 0, false, payload)
}

func (c *Controller) publishState(f fileData, payload []byte) error {
	topic := c.stateTopic(f.sensorID)
	return c.mqtt.Publish(topic, 0, false, payload)
}

func (c *Controller) configTopic(component, sensorID string) string {
	return fmt.Sprintf("%s/%s/%s/%s/config", c.settings.DiscoveryPrefix, component, c.settings.DeviceID, sensorID)
}

func (c *Controller) stateTopic(sensorID string) string {
	return fmt.Sprintf("%s/%s/state", c.settings.StatePrefix, sensorID)
}

func validateFile(f fileData) bool {
	if f.data.Component == "" {
		log.Println("invalid config.component in", f.name)
		return false
	}
	if f.data.Config["name"] == "" || f.data.Config["name"] == nil {
		log.Println("invalid config.component in", f.name)
		return false
	}
	if f.data.Config["device_class"] == "" || f.data.Config["device_class"] == nil {
		log.Println("invalid config.component in", f.name)
		return false
	}

	return true
}
