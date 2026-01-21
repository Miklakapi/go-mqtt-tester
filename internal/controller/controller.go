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

func (c *Controller) Init() error {
	entries, err := os.ReadDir(c.settings.DataDir)
	if err != nil {
		return fmt.Errorf("%w: %s: %w", ErrReadDir, c.settings.DataDir, err)
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !strings.HasSuffix(e.Name(), ".json") {
			continue
		}

		f, err := c.getFileData(e.Name())
		if err != nil {
			log.Println(err)
			continue
		}

		if err := validateFile(f); err != nil {
			log.Println(err)
			continue
		}

		configPayload, err := json.Marshal(f.data.Config)
		if err != nil {
			log.Println("config marshal error:", err)
			continue
		}

		statePayload, err := buildStatePayload(f)
		if err != nil {
			log.Println(err)
			continue
		}

		if err := c.publishConfig(f, configPayload); err != nil {
			log.Println("publish config error:", err)
		}
		if err := c.publishState(f, statePayload); err != nil {
			log.Println("publish state error:", err)
		}
	}

	return nil
}

func (c *Controller) Run(ctx context.Context) error {
	c.watcher.AddWatch(c.settings.DataDir, syscall.IN_DELETE|syscall.IN_MOVED_FROM|syscall.IN_MOVED_TO|syscall.IN_CLOSE_WRITE)

	for {
		select {
		case event := <-c.watcher.Events:
			if event.Mask&syscall.IN_ISDIR != 0 {
				continue
			}
			fileData, err := c.getFileData(event.Name)
			if err != nil {
				log.Println(err)
				continue
			}

			if err := validateFile(fileData); err != nil {
				log.Println(err)
				continue
			}

			configPayload, err := json.Marshal(fileData.data.Config)
			if err != nil {
				log.Println("config marshal error:", err)
				continue
			}

			statePayload, err := buildStatePayload(fileData)
			if err != nil {
				log.Println(err)
				continue
			}

			if err := c.publishConfig(fileData, configPayload); err != nil {
				log.Println("publish config error:", err)
			}
			if err := c.publishState(fileData, statePayload); err != nil {
				log.Println("publish state error:", err)
			}

		case err := <-c.watcher.Errors:
			log.Println(err)

		case <-ctx.Done():
			return nil
		}
	}
}

func (c *Controller) getFileData(name string) (fileData, error) {
	sensorID := strings.TrimSuffix(name, ".json")
	path := filepath.Join(c.settings.DataDir, name)

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
		name:     name,
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

func validateFile(f fileData) error {
	if strings.TrimSpace(f.data.Component) == "" {
		return fmt.Errorf("%w: %s", ErrMissingComponent, f.name)
	}

	name, ok := f.data.Config["name"]
	if !ok || name == nil || strings.TrimSpace(fmt.Sprint(name)) == "" {
		return fmt.Errorf("%w: %s", ErrMissingName, f.name)
	}

	deviceClass, ok := f.data.Config["device_class"]
	if !ok || deviceClass == nil || strings.TrimSpace(fmt.Sprint(deviceClass)) == "" {
		return fmt.Errorf("%w: %s", ErrMissingDeviceClass, f.name)
	}

	if f.data.Component == "binary_sensor" {
		s, ok := f.data.State.(string)
		if !ok || (s != "ON" && s != "OFF") {
			return fmt.Errorf("%w: %s", ErrInvalidBinarySensorState, f.name)
		}
	}

	return nil
}

func buildStatePayload(f fileData) ([]byte, error) {
	if f.data.Component == "binary_sensor" {
		return []byte(f.data.State.(string)), nil
	}

	b, err := json.Marshal(f.data.State)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %w", ErrInvalidJSON, f.name, err)
	}

	return b, nil
}
