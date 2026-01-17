package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	MqttBrokerHost string
	MqttBrokerPort string
	MqttUsername   string
	MqttPassword   string
	MqttClientId   string

	MqttDiscoveryPrefix string
	MqttStatePrefix     string

	HaDeviceId           string
	HaDeviceName         string
	HaDeviceManufacturer string
	HaDeviceModel        string

	DataDir string
}

func Load() (Config, error) {
	cfg := Config{}

	var err error

	cfg.MqttBrokerHost, err = getEnvRequired("MQTT_BROKER_HOST")
	if err != nil {
		return Config{}, err
	}
	cfg.MqttBrokerPort = getEnvDefault("MQTT_BROKER_PORT", "1883")
	cfg.MqttUsername, err = getEnvRequired("MQTT_USERNAME")
	if err != nil {
		return Config{}, err
	}
	cfg.MqttPassword, err = getEnvRequired("MQTT_PASSWORD")
	if err != nil {
		return Config{}, err
	}
	cfg.MqttClientId = getEnvDefault("MQTT_CLIENT_ID", "go-mqtt-tester")

	cfg.MqttDiscoveryPrefix = getEnvDefault("MQTT_DISCOVERY_PREFIX", "homeassistant")
	cfg.MqttStatePrefix = getEnvDefault("MQTT_STATE_PREFIX", "go-mqtt-tester")

	cfg.HaDeviceId = getEnvDefault("HA_DEVICE_ID", "go-mqtt-tester")
	cfg.HaDeviceName = getEnvDefault("HA_DEVICE_NAME", "Go MQTT Tester")
	cfg.HaDeviceManufacturer = getEnvDefault("HA_DEVICE_MANUFACTURER", "Local Dev")
	cfg.HaDeviceModel = getEnvDefault("HA_DEVICE_MODEL", "Fake IoT Device")

	cfg.DataDir = getEnvDefault("DATA_DIR", "./data")

	return cfg, nil
}

func getEnvRequired(key string) (string, error) {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return "", fmt.Errorf("configuration error: required environment variable %q is not set", key)
	}
	return val, nil
}

func getEnvDefault(key, def string) string {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return def
	}
	return val
}
