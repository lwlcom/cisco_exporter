package main

import (
	"os"
	"strings"

	"github.com/lwlcom/cisco_exporter/config"
	"github.com/lwlcom/cisco_exporter/connector"
	"github.com/pkg/errors"
)

func devicesForConfig(cfg *config.Config) ([]*connector.Device, error) {
	devs := make([]*connector.Device, len(cfg.Devices))
	var err error
	for i, d := range cfg.Devices {
		devs[i], err = deviceFromDeviceConfig(d, cfg)
		if err != nil {
			return nil, err
		}
	}

	return devs, nil
}

func deviceFromDeviceConfig(device *config.DeviceConfig, cfg *config.Config) (*connector.Device, error) {
	auth, err := authForDevice(device, cfg)
	if err != nil {
		return nil, errors.Wrapf(err, "could not initialize config for device %s", device.Host)
	}

	port := "22"
	host := device.Host
	if strings.Contains(host, ":") {
		d := strings.Split(host, ":")
		host = d[0]
		port = d[1]
	}

	return &connector.Device{
		Host:         host,
		Port:         port,
		Auth:         auth,
		DeviceConfig: device,
	}, nil
}

func authForDevice(device *config.DeviceConfig, cfg *config.Config) (connector.AuthMethod, error) {
	user := cfg.Username
	if device.Username != nil {
		user = *device.Username
	}

	if device.KeyFile != nil {
		return authForKeyFile(user, *device.KeyFile)
	}

	if cfg.KeyFile != "" {
		return authForKeyFile(user, cfg.KeyFile)
	}

	if device.Password != nil {
		return connector.AuthByPassword(user, *device.Password), nil
	}

	if cfg.Password != "" {
		return connector.AuthByPassword(user, cfg.Password), nil
	}

	return nil, errors.New("no valid authentication method available")
}

func authForKeyFile(username, keyFile string) (connector.AuthMethod, error) {
	f, err := os.Open(keyFile)
	if err != nil {
		return nil, errors.Wrap(err, "could not open ssh key file")
	}
	defer f.Close()
	auth, err := connector.AuthByKey(username, f)
	if err != nil {
		return nil, errors.Wrap(err, "could not load ssh private key file")
	}

	return auth, nil
}
