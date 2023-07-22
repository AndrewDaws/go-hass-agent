// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package agent

import (
	"errors"
	"fmt"
	"net/url"

	"fyne.io/fyne/v2"
	"github.com/go-playground/validator/v10"
)

const (
	websocketPath    = "/api/websocket"
	webHookPath      = "/api/webhook/"
	PrefApiURL       = "ApiURL"
	PrefWebsocketURL = "WebSocketURL"
	PrefToken        = "Token"
	PrefWebhookID    = "WebhookID"
	PrefSecret       = "secret"
)

type config interface {
	Get(string) (string, error)
	Set(string, string) error
}

type agentConfig struct {
	prefs fyne.Preferences
}

func (agent *Agent) LoadConfig() *agentConfig {
	return &agentConfig{
		prefs: agent.app.Preferences(),
	}
}

func (c *agentConfig) WebSocketURL() string {
	return c.prefs.String(PrefWebsocketURL)
}

func (c *agentConfig) WebhookID() string {
	return c.prefs.String(PrefWebhookID)
}

func (c *agentConfig) Token() string {
	return c.prefs.String(PrefToken)
}

func (c *agentConfig) ApiURL() string {
	return c.prefs.String(PrefApiURL)
}

func (c *agentConfig) Secret() string {
	return c.prefs.String(PrefSecret)
}

func (c *agentConfig) NewStorage(id string) (string, error) {
	registryPath, err := extraStoragePath(id)
	if err != nil {
		return "", err
	}
	return registryPath.Path(), nil
}

func (c *agentConfig) Get(key string) (string, error) {
	value := c.prefs.StringWithFallback(key, "NOTSET")
	if value == "NOTSET" {
		return "", errors.New("key not set")
	}
	return value, nil
}

func (c *agentConfig) Set(key, value string) error {
	c.prefs.SetString(key, value)
	return nil
}

func ValidateConfig(c config) error {
	validator := validator.New()

	validate := func(key, rules, errMsg string) error {
		value, err := c.Get(key)
		if err != nil {
			return fmt.Errorf("unable to retrieve %s from config", key)
		}
		err = validator.Var(value, rules)
		if err != nil {
			return errors.New(errMsg)
		}
		return nil
	}

	if err := validate(PrefApiURL,
		"required,url",
		"apiURL does not match either a URL, hostname or hostname:port",
	); err != nil {
		return err
	}
	if err := validate(PrefWebsocketURL,
		"required,url",
		"websocketURL does not match either a URL, hostname or hostname:port",
	); err != nil {
		return err
	}
	if err := validate(PrefToken,
		"required,ascii",
		"invalid long-lived token format",
	); err != nil {
		return err
	}
	if err := validate(PrefWebhookID,
		"required,ascii",
		"invalid webhookID format",
	); err != nil {
		return err
	}

	return nil
}

func (c *agentConfig) generateWebsocketURL() {
	// TODO: look into websocket http upgrade method
	host := c.prefs.String("Host")
	url, _ := url.Parse(host)
	switch url.Scheme {
	case "https":
		url.Scheme = "wss"
	case "http":
		fallthrough
	default:
		url.Scheme = "ws"
	}
	url = url.JoinPath(websocketPath)
	c.prefs.SetString(PrefWebsocketURL, url.String())
}

func (c *agentConfig) generateAPIURL() {
	cloudhookURL := c.prefs.String("CloudhookURL")
	remoteUIURL := c.prefs.String("RemoteUIURL")
	webhookID := c.prefs.String(PrefWebhookID)
	host := c.prefs.String("Host")
	var apiURL string
	switch {
	case cloudhookURL != "":
		apiURL = cloudhookURL
	case remoteUIURL != "" && webhookID != "":
		apiURL = remoteUIURL + webHookPath + webhookID
	case webhookID != "" && host != "":
		url, _ := url.Parse(host)
		url = url.JoinPath(webHookPath, webhookID)
		apiURL = url.String()
	default:
		apiURL = ""
	}
	c.prefs.SetString("ApiURL", apiURL)
}
