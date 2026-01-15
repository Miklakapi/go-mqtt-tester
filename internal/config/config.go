package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Mode string
}

func Load() (Config, error) {
	cfg := Config{}

	cfg.Mode = getEnv("MODE", "dev")

	return cfg, nil
}

func getEnv(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}

func parseIntEnv(key string, def int) (int, error) {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return def, nil
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	}
	return n, nil
}

func parseDurationEnv(key, def string) (time.Duration, error) {
	val := getEnv(key, def)
	d, err := time.ParseDuration(val)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	}
	return d, nil
}
