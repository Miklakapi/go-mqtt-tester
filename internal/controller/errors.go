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

	ErrReadFile    = errors.New("failed to read data file")
	ErrInvalidJSON = errors.New("invalid json file")
)
