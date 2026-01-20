# Go-File-Share

![license](https://img.shields.io/badge/license-MIT-blue)
![linux](https://img.shields.io/badge/os-Linux-green)
![language](https://img.shields.io/badge/language-Go_1.25.1-blue)
![version](https://img.shields.io/badge/version-1.0.0-success)
![status](https://img.shields.io/badge/status-development-blue)

A small Go application for testing MQTT-based sensors in Home Assistant.
The project allows defining sensors using simple JSON files and publishing their configuration and state to Home Assistant via MQTT.

This project was created mainly as a playground for learning MQTT, Home Assistant discovery, and low-level file system monitoring rather than as a production-ready tool.

## Table of Contents

- [General info](#general-info)
- [How it works](#how-it-works)
- [Architecture](#architecture)
- [Technologies](#technologies)
- [Setup](#setup)
- [Data format](#data-format)
- [Features](#features)
- [Status](#status)

## General info

Go-MQTT-Tester is a small utility that lets you quickly prototype and test Home Assistant sensors using MQTT.

Instead of writing YAML or ESPHome configs, you define sensors as JSON files in a directory.
The application watches that directory and automatically:

- publishes MQTT Discovery configuration,
- publishes sensor state updates,
- updates Home Assistant in real time when files change.

Typical use cases:

- testing MQTT discovery payloads,
- experimenting with different sensor types (sensor, binary_sensor, etc.),
- learning how Home Assistant reacts to MQTT messages,
- validating state formats without flashing devices.

The application is intentionally stateless:

> All sensor configuration and state comes exclusively from files.
> Restarting the application simply republishes everything from disk.

## How it works

1. The application watches a directory (e.g. ./data) using Linux inotify
2. Each .json file represents one Home Assistant entity
3. On startup:

- all files are read
- configuration and state are published to MQTT

4. On file change:

- configuration is re-published
- state is updated immediately

5. On Home Assistant side:

- entities appear automatically via MQTT Discovery

No database.
No persistence.
No background magic.

## Architecture

The project intentionally keeps the architecture simple and explicit, without unnecessary abstractions.

Core components:

- Watcher
    - custom inotify-based file watcher
    - no external fsnotify dependency

- Controller
    - connects file events with MQTT publishing
    - validates input data
    - enriches configs with device metadata

- MQTT client
    - thin wrapper over paho.mqtt.golang
    - only Publish and Close

There are no interfaces, no dependency injection frameworks, and no adapters.
This is a conscious decision — the project has one job and one implementation.

## Technologies

Project is created with:

- Go 1.25.1
- Linux inotify (syscall)
- MQTT:
    - github.com/eclipse/paho.mqtt.golang v1.5.1
- Home Assistant MQTT Discovery
- JSON-based configuration
- Plain Go standard library

## Setup

### Requirements

- Linux (inotify required)
- Running MQTT broker (e.g. Mosquitto)
- Home Assistant with MQTT integration enabled

### Application

1. Install dependencies:

```go
go mod tidy
```

2. Configure environment variables (example):

```ini
MQTT_BROKER_HOST=localhost
MQTT_BROKER_PORT=1883
MQTT_USERNAME=homeassistant
MQTT_PASSWORD=secret

MQTT_DISCOVERY_PREFIX=homeassistant
MQTT_STATE_PREFIX=go-mqtt-tester

HA_DEVICE_ID=go_mqtt_tester
HA_DEVICE_NAME=Go MQTT Tester
HA_DEVICE_MANUFACTURER=Local Dev
HA_DEVICE_MODEL=Virtual Device

DATA_DIR=./data
```

3. Run the application:

```go
go run ./cmd/main.go
```

## Data format

Each sensor is defined by a single JSON file:

```json
{
    "component": "sensor",
    "config": {
        "name": "Outside Temperature",
        "device_class": "temperature",
        "unit_of_measurement": "°C"
    },
    "state": 23.7
}
```

Notes:

- component determines the Home Assistant entity type
- state format depends on the component
    - binary_sensor: "ON" / "OFF"
    - others: any valid JSON value
- unique_id, state_topic and device are automatically generated

## Features

- MQTT Discovery support
- Automatic entity creation in Home Assistant
- Live updates on file change
- Manual sensor state control via files
- Supports multiple sensor types
- No database, no persistence
- Custom inotify-based file watcher
- Minimal MQTT wrapper
- Designed for experimentation and learning

## Status

The project is in active development.
