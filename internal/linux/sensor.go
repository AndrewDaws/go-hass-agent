// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package linux

import (
	"github.com/iancoleman/strcase"
	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
)

const (
	srcDbus   = "D-Bus"
	srcProcfs = "ProcFS"
)

// linuxSensor represents a generic sensor on the Linux platform. Most sensors
// will be able to use this struct, which satisfies the tracker.Sensor
// interface, alllowing them to be sent as a sensor to Home Assistant.
type linuxSensor struct {
	value  any
	icon   string
	units  string
	source string
	sensorType
	isBinary     bool
	isDiagnostic bool
	deviceClass  sensor.SensorDeviceClass
	stateClass   sensor.SensorStateClass
}

// linuxSensor satisfies the tracker.Sensor interface, allowing it to be sent as
// a sensor update to Home Assistant. Any of the methods below can be overridden
// by embedding linuxSensor in another struct and defining the needed function.

func (l *linuxSensor) Name() string {
	return l.sensorType.String()
}

func (l *linuxSensor) ID() string {
	return strcase.ToSnake(l.sensorType.String())
}

func (l *linuxSensor) State() any {
	return l.value
}

func (l *linuxSensor) SensorType() sensor.SensorType {
	if l.isBinary {
		return sensor.TypeBinary
	}
	return sensor.TypeSensor
}

func (l *linuxSensor) Category() string {
	if l.isDiagnostic {
		return "diagnostic"
	}
	return ""
}

func (l *linuxSensor) DeviceClass() sensor.SensorDeviceClass {
	return l.deviceClass
}

func (l *linuxSensor) StateClass() sensor.SensorStateClass {
	return l.stateClass
}

func (l *linuxSensor) Icon() string {
	return l.icon
}

func (l *linuxSensor) Units() string {
	return l.units
}

func (l *linuxSensor) Attributes() any {
	if l.source != "" {
		return struct {
			DataSource string `json:"Data Source"`
		}{
			DataSource: l.source,
		}
	}
	return nil
}
