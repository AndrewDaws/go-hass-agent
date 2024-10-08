// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"context"
	"errors"
	"log/slog"

	mqtthass "github.com/joshuar/go-hass-anything/v11/pkg/hass"
	mqttapi "github.com/joshuar/go-hass-anything/v11/pkg/mqtt"

	"github.com/joshuar/go-hass-agent/internal/linux"
	"github.com/joshuar/go-hass-agent/internal/linux/apps"
	"github.com/joshuar/go-hass-agent/internal/linux/battery"
	"github.com/joshuar/go-hass-agent/internal/linux/cpu"
	"github.com/joshuar/go-hass-agent/internal/linux/desktop"
	"github.com/joshuar/go-hass-agent/internal/linux/disk"
	"github.com/joshuar/go-hass-agent/internal/linux/location"
	"github.com/joshuar/go-hass-agent/internal/linux/media"
	"github.com/joshuar/go-hass-agent/internal/linux/mem"
	"github.com/joshuar/go-hass-agent/internal/linux/net"
	"github.com/joshuar/go-hass-agent/internal/linux/power"
	"github.com/joshuar/go-hass-agent/internal/linux/problems"
	"github.com/joshuar/go-hass-agent/internal/linux/system"
	"github.com/joshuar/go-hass-agent/internal/linux/user"
	"github.com/joshuar/go-hass-agent/internal/logging"
)

// allworkers is the list of sensor allworkers supported on Linux.
var allworkers = []func(context.Context) (*linux.SensorWorker, error){
	apps.NewAppWorker,
	battery.NewBatteryWorker,
	cpu.NewUsageWorker,
	cpu.NewLoadAvgWorker,
	cpu.NewUsageWorker,
	desktop.NewDesktopWorker,
	disk.NewIOWorker,
	disk.NewUsageWorker,
	location.NewLocationWorker,
	mem.NewUsageWorker,
	net.NewConnectionWorker,
	net.NewRatesWorker,
	power.NewLaptopWorker,
	power.NewProfileWorker,
	power.NewStateWorker,
	power.NewScreenLockWorker,
	problems.NewProblemsWorker,
	system.NewHWMonWorker,
	system.NewInfoWorker,
	system.NewTimeWorker,
	user.NewUserWorker,
}

var (
	ErrWorkerAlreadyStarted = errors.New("worker already started")
	ErrUnknownWorker        = errors.New("unknown worker")
)

type mqttWorker struct {
	msgs          chan *mqttapi.Msg
	sensors       []*mqtthass.SensorEntity
	buttons       []*mqtthass.ButtonEntity
	numbers       []*mqtthass.NumberEntity[int]
	switches      []*mqtthass.SwitchEntity
	controls      []*mqttapi.Subscription
	binarySensors []*mqtthass.BinarySensorEntity
	cameras       []*mqtthass.ImageEntity
}

type linuxSensorController struct {
	deviceController
}

type linuxMQTTController struct {
	*mqttWorker
	logger *slog.Logger
}

// entity is a convienience interface to avoid duplicating a lot of loop content
// when configuring the controller.
type entity interface {
	MarshalSubscription() (*mqttapi.Subscription, error)
	MarshalConfig() (*mqttapi.Msg, error)
}

func (c *linuxMQTTController) Subscriptions() []*mqttapi.Subscription {
	totalLength := len(c.buttons) + len(c.numbers) + len(c.switches) + len(c.cameras)
	subs := make([]*mqttapi.Subscription, 0, totalLength)

	// Create subscriptions for buttons.
	for _, button := range c.buttons {
		subs = append(subs, c.generateSubscription(button))
	}
	// Create subscriptions for numbers.
	for _, number := range c.numbers {
		subs = append(subs, c.generateSubscription(number))
	}
	// Create subscriptions for switches.
	for _, sw := range c.switches {
		subs = append(subs, c.generateSubscription(sw))
	}
	// Add subscriptions for any additional controls.
	subs = append(subs, c.controls...)

	return subs
}

func (c *linuxMQTTController) Configs() []*mqttapi.Msg {
	totalLength := len(c.sensors) + len(c.binarySensors) + len(c.buttons) + len(c.switches) + len(c.numbers) + len(c.cameras)
	configs := make([]*mqttapi.Msg, 0, totalLength)

	// Create sensor configs.
	for _, sensorEntity := range c.sensors {
		configs = append(configs, c.generateConfig(sensorEntity))
	}
	// Create binary sensor configs.
	for _, binarySensorEntity := range c.binarySensors {
		configs = append(configs, c.generateConfig(binarySensorEntity))
	}
	// Create button configs.
	for _, buttonEntity := range c.buttons {
		configs = append(configs, c.generateConfig(buttonEntity))
	}
	// Create number configs.
	for _, numberEntity := range c.numbers {
		configs = append(configs, c.generateConfig(numberEntity))
	}
	// Create switch configs.
	for _, switchEntity := range c.switches {
		configs = append(configs, c.generateConfig(switchEntity))
	}
	// Create camera configs.
	for _, cameraEntity := range c.cameras {
		configs = append(configs, c.generateConfig(cameraEntity))
	}

	return configs
}

func (c *linuxMQTTController) Msgs() chan *mqttapi.Msg {
	return c.msgs
}

// generateConfig is a helper function to avoid duplicate code around generating
// an entity subscription.
func (c *linuxMQTTController) generateSubscription(e entity) *mqttapi.Subscription {
	sub, err := e.MarshalSubscription()
	if err != nil {
		c.logger.Warn("Could not create subscription.", slog.Any("error", err))

		return nil
	}

	return sub
}

// generateConfig is a helper function to avoid duplicate code around generating
// an entity config.
func (c *linuxMQTTController) generateConfig(e entity) *mqttapi.Msg {
	cfg, err := e.MarshalConfig()
	if err != nil {
		c.logger.Warn("Could not create config.", slog.Any("error", err.Error()))

		return nil
	}

	return cfg
}

// newOSController initializes the list of workers for sensors and returns those
// that are supported on this device.
//
//revive:disable:function-length
//nolint:cyclop
func (agent *Agent) newOSController(ctx context.Context, mqttDevice *mqtthass.Device) (SensorController, MQTTController) {
	ctx = linux.NewContext(ctx)

	logger := agent.logger.With(slog.Group("linux", slog.String("controller", "sensor")))
	ctx = logging.ToContext(ctx, logger)
	sensorController := &linuxSensorController{
		deviceController: deviceController{
			sensorWorkers: make(map[string]*sensorWorker),
			logger:        logger,
		},
	}

	// Set up sensor workers.
	for _, startWorkerFunc := range allworkers {
		worker, err := startWorkerFunc(ctx)
		if err != nil {
			sensorController.logger.Warn("Could not start a sensor worker.", slog.Any("error", err))

			continue
		}

		sensorController.sensorWorkers[worker.ID()] = &sensorWorker{object: worker, started: false}
	}

	// Stop setup if there is no mqttDevice.
	if mqttDevice == nil {
		return sensorController, nil
	}

	logger = agent.logger.With(slog.Group("linux", slog.String("controller", "mqtt")))
	ctx = logging.ToContext(ctx, logger)
	mqttController := &linuxMQTTController{
		mqttWorker: &mqttWorker{
			msgs: make(chan *mqttapi.Msg),
		},
		logger: logger,
	}

	// Add the power controls (suspend, resume, poweroff, etc.).
	powerEntities, err := power.NewPowerControl(ctx, mqttDevice)
	if err != nil {
		mqttController.logger.Warn("Could not create power controls.", slog.Any("error", err))
	} else {
		mqttController.buttons = append(mqttController.buttons, powerEntities...)
	}
	// Add the screen lock controls.
	screenControls, err := power.NewScreenLockControl(ctx, mqttDevice)
	if err != nil {
		mqttController.logger.Warn("Could not create screen lock controls.", slog.Any("error", err))
	} else {
		mqttController.buttons = append(mqttController.buttons, screenControls...)
	}
	// Add the volume controls.
	volEntity, muteEntity := media.VolumeControl(ctx, mqttController.Msgs(), mqttDevice)
	if volEntity != nil && muteEntity != nil {
		mqttController.numbers = append(mqttController.numbers, volEntity)
		mqttController.switches = append(mqttController.switches, muteEntity)
	}
	// Add media control.
	mprisEntity, err := media.MPRISControl(ctx, mqttDevice, mqttController.Msgs())
	if err != nil {
		mqttController.logger.Warn("Could not activate MPRIS controller.", slog.Any("error", err))
	} else {
		mqttController.sensors = append(mqttController.sensors, mprisEntity)
	}
	// Add camera control.
	cameraEntities := media.NewCameraControl(ctx, mqttController.Msgs(), mqttDevice)
	if cameraEntities != nil {
		mqttController.buttons = append(mqttController.buttons, cameraEntities.StartButton, cameraEntities.StopButton)
		mqttController.cameras = append(mqttController.cameras, cameraEntities.Images)
		mqttController.sensors = append(mqttController.sensors, cameraEntities.Status)
	}

	// Add the D-Bus command action.
	dbusCmdController, err := system.NewDBusCommandSubscription(ctx, mqttDevice)
	if err != nil {
		mqttController.logger.Warn("Could not activate D-Bus commands controller.", slog.Any("error", err))
	} else {
		mqttController.controls = append(mqttController.controls, dbusCmdController)
	}

	go func() {
		defer close(mqttController.msgs)
		<-ctx.Done()
	}()

	return sensorController, mqttController
}
