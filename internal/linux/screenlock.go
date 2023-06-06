// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"context"

	"github.com/godbus/dbus/v5"
	"github.com/joshuar/go-hass-agent/internal/device"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/rs/zerolog/log"
)

const (
	screensaverDBusPath      = "/org/freedesktop/ScreenSaver"
	screensaverDBusInterface = "org.freedesktop.ScreenSaver"
)

type screenlock struct {
	locked bool
}

func (l *screenlock) Name() string {
	return "Screen Lock"
}

func (l *screenlock) ID() string {
	return "screen_lock"
}

func (l *screenlock) Icon() string {
	if l.locked {
		return "mdi:eye-lock"
	} else {
		return "mdi:eye-lock-open"
	}
}

func (l *screenlock) SensorType() hass.SensorType {
	return hass.TypeBinary
}

func (l *screenlock) DeviceClass() hass.SensorDeviceClass {
	return 0
}

func (l *screenlock) StateClass() hass.SensorStateClass {
	return 0
}

func (l *screenlock) State() interface{} {
	return l.locked
}

func (l *screenlock) Units() string {
	return ""
}

func (l *screenlock) Category() string {
	return ""
}

func (l *screenlock) Attributes() interface{} {
	return nil
}

func ScreenLockUpdater(ctx context.Context, update chan interface{}) {
	deviceAPI, err := device.FetchAPIFromContext(ctx)
	if err != nil {
		log.Debug().Err(err).Caller().
			Msg("Could not connect to DBus.")
		return
	}
	dbusAPI := device.GetAPIEndpoint[*bus](deviceAPI, "session")

	NewBusRequest(dbusAPI).
		Path(screensaverDBusPath).
		Match([]dbus.MatchOption{
			dbus.WithMatchObjectPath(screensaverDBusPath),
			dbus.WithMatchInterface(screensaverDBusInterface),
		}).
		Event("org.freedesktop.ScreenSaver.ActiveChanged").
		Handler(func(s *dbus.Signal) {
			lock := new(screenlock)
			lock.locked = s.Body[0].(bool)
			update <- lock
		}).
		AddWatch(ctx)
}
