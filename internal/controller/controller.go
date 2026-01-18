package controller

import (
	"context"

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
	<-ctx.Done()
	return nil
}
