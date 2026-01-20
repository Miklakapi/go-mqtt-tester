package controller

import "errors"

var (
	ErrMissingDiscoveryPrefix    = errors.New("missing discovery prefix")
	ErrMissingStatePrefix        = errors.New("missing state prefix")
	ErrMissingDeviceID           = errors.New("missing device id")
	ErrMissingDeviceName         = errors.New("missing device name")
	ErrMissingDeviceManufacturer = errors.New("missing device manufacturer")
	ErrMissingDeviceModel        = errors.New("missing device model")
	ErrMissingDataDir            = errors.New("missing data dir")

	ErrMissingComponent         = errors.New("missing component")
	ErrMissingName              = errors.New("missing config.name")
	ErrMissingDeviceClass       = errors.New("missing config.device_class")
	ErrInvalidBinarySensorState = errors.New("invalid binary_sensor state (expected ON/OFF)")

	ErrReadFile    = errors.New("failed to read data file")
	ErrReadDir     = errors.New("read dir failed")
	ErrInvalidJSON = errors.New("invalid json file")
)
