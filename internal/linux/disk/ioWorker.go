// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//revive:disable:unused-receiver
package disk

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/joshuar/go-hass-agent/internal/hass/sensor"
	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/logging"
)

const (
	ratesUpdateInterval = 5 * time.Second
	ratesUpdateJitter   = time.Second

	ratesWorkerID = "disk_rates_sensors"
)

// ioWorker creates sensors for disk IO counts and rates per device. It
// maintains an internal map of devices being tracked.
type ioWorker struct {
	boottime time.Time
	devices  map[string][]*diskIOSensor
	mu       sync.Mutex
}

// addDevice adds a new device to the tracker map. If sthe device is already
// being tracked, it will not be added again. The bool return indicates whether
// a device was added (true) or not (false).
func (w *ioWorker) addDevice(dev *device) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, found := w.devices[dev.id]; !found {
		w.devices[dev.id] = newDeviceSensors(w.boottime, dev)
	}
}

// updateDevice will update a tracked device's stats. For rates, it will
// recalculate based on the given time delta.
func (w *ioWorker) updateDevice(dev *device, stats map[stat]uint64, delta time.Duration) []sensor.Details {
	w.mu.Lock()
	defer w.mu.Unlock()

	sensors := make([]sensor.Details, len(w.devices["total"]))

	if _, found := w.devices[dev.id]; found && stats != nil {
		for idx := range w.devices[dev.id] {
			w.devices[dev.id][idx].update(stats, delta)
			sensors[idx] = w.devices[dev.id][idx]
		}
	}

	return sensors
}

func (w *ioWorker) updateTotals(stats map[stat]uint64, delta time.Duration) []sensor.Details {
	w.mu.Lock()
	defer w.mu.Unlock()

	sensors := make([]sensor.Details, len(w.devices["total"]))

	for idx := range w.devices["total"] {
		w.devices["total"][idx].update(stats, delta)
		sensors[idx] = w.devices["total"][idx]
	}

	return sensors
}

func (w *ioWorker) Interval() time.Duration { return ratesUpdateInterval }

func (w *ioWorker) Jitter() time.Duration { return ratesUpdateJitter }

func (w *ioWorker) Sensors(ctx context.Context, duration time.Duration) ([]sensor.Details, error) {
	// Get valid devices.
	deviceNames, err := getDeviceNames()
	if err != nil {
		return nil, fmt.Errorf("could not fetch disk devices: %w", err)
	}

	sensors := make([]sensor.Details, 0, 4*len(deviceNames)+4) //nolint:mnd
	totals := make(map[stat]uint64)

	// Get the current device info and stats for all valid devices.
	for _, name := range deviceNames {
		dev, stats, err := getDevice(name)
		if err != nil {
			logging.FromContext(ctx).
				With(slog.String("worker", ratesWorkerID)).
				Debug("Unable to read device stats.", slog.Any("error", err))

			continue
		}

		// Add device (if it isn't already tracked).
		w.addDevice(dev)

		// Update device stats and return updated sensors.
		sensors = append(sensors, w.updateDevice(dev, stats, duration)...)

		// Don't include "aggregate" devices in totals.
		if strings.HasPrefix(dev.id, "dm") || strings.HasPrefix(dev.id, "md") {
			continue
		}
		// Add device stats to the totals.
		for stat, value := range stats {
			totals[stat] += value
		}
	}

	// Update total stats and return updated sensors.
	sensors = append(sensors, w.updateTotals(totals, duration)...)

	return sensors, nil
}

func NewIOWorker(ctx context.Context) (*linux.SensorWorker, error) {
	boottime, found := linux.CtxGetBoottime(ctx)
	if !found {
		return nil, fmt.Errorf("%w: no boottime value", linux.ErrInvalidCtx)
	}

	worker := &ioWorker{
		devices:  make(map[string][]*diskIOSensor),
		boottime: boottime,
	}

	// Add sensors for a pseudo "total" device which tracks total values from
	// all devices.
	worker.devices["total"] = newDeviceSensors(worker.boottime, &device{id: "total"})

	return &linux.SensorWorker{
			Value:    worker,
			WorkerID: ratesWorkerID,
		},
		nil
}

func newDeviceSensors(boottime time.Time, dev *device) []*diskIOSensor {
	return []*diskIOSensor{
		newDiskIOSensor(boottime, dev, diskReads),
		newDiskIOSensor(boottime, dev, diskWrites),
		newDiskIORateSensor(dev, diskReadRate),
		newDiskIORateSensor(dev, diskWriteRate),
	}
}
