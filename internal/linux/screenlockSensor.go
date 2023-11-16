// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"

	"github.com/godbus/dbus/v5"
	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/rs/zerolog/log"
)

const (
	screensaverDBusPath      = "/org/freedesktop/ScreenSaver"
	screensaverDBusInterface = "org.freedesktop.ScreenSaver"
)

type screenlockSensor struct {
	linuxSensor
}

func (s *screenlockSensor) Icon() string {
	if s.value.(bool) {
		return "mdi:eye-lock"
	} else {
		return "mdi:eye-lock-open"
	}
}

func (s *screenlockSensor) SensorType() sensor.SensorType {
	return sensor.TypeBinary
}

func newScreenlockEvent(v bool) *screenlockSensor {
	return &screenlockSensor{
		linuxSensor: linuxSensor{
			sensorType: screenLock,
			source:     srcDbus,
			value:      v,
		},
	}
}

func ScreenLockUpdater(ctx context.Context, tracker device.SensorTracker) {
	path := getSessionPath(ctx)
	if path == "" {
		log.Warn().Msg("Could not ascertain user session from D-Bus. Cannot monitor screen lock state.")
		return
	}
	err := NewBusRequest(ctx, SystemBus).
		Path(path).
		Match([]dbus.MatchOption{
			dbus.WithMatchPathNamespace("/org/freedesktop/login1"),
			dbus.WithMatchInterface("org.freedesktop.DBus.Properties"),
		}).
		Event("org.freedesktop.DBus.Properties.PropertiesChanged").
		Handler(func(s *dbus.Signal) {
			props, ok := s.Body[1].(map[string]dbus.Variant)
			if !ok {
				log.Warn().Str("signal", s.Name).Interface("body", s.Body).
					Msg("Unexpected signal body")
				return
			}
			if v, ok := props["LockedHint"]; ok {
				lock := newScreenlockEvent(variantToValue[bool](v))
				if err := tracker.UpdateSensors(ctx, lock); err != nil {
					log.Error().Err(err).Msg("Could not update screen lock sensor.")
				}
			}
		}).
		AddWatch(ctx)
	if err != nil {
		log.Warn().Err(err).
			Msg("Could not poll D-Bus for screen lock. Screen lock sensor will not run.")
	}
}
