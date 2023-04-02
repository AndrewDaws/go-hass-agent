package agent

import (
	"context"

	"fyne.io/fyne/v2"
	"github.com/joshuar/go-hass-agent/internal/config"
	"github.com/joshuar/go-hass-agent/internal/hass"
	"github.com/rs/zerolog/log"
)

func (agent *Agent) runNotificationsWorker(ctx context.Context) {
	for {
		ws := hass.NewWebsocket(ctx)
		if ws == nil {
			log.Debug().Caller().
				Msgf("No websocket connection made.")
			return
		} else {
			go agent.handleNotifications(ctx, ws.ReadCh, ws.WriteCh)
		}
		select {
		case <-ctx.Done():
			log.Debug().Caller().Msg("Closing notifications worker.")
			ws.Close()
			return
		}
	}
}

func (agent *Agent) handleNotifications(ctx context.Context, response chan *hass.WebsocketResponse, request chan interface{}) {

	config, validConfig := config.FromContext(ctx)
	if !validConfig {
		log.Debug().Caller().Msg("Could not retrieve valid config from context.")
		return
	}

	for {
		select {
		case <-ctx.Done():
			log.Debug().Caller().Msg("Stopping handling notifications.")
			return
		case r := <-response:
			switch r.Type {
			case "auth_required":
				log.Debug().Caller().Msg("Requesting authorisation for websocket.")
				request <- struct {
					Type        string `json:"type"`
					AccessToken string `json:"access_token"`
				}{
					Type:        "auth",
					AccessToken: config.Token,
				}
			case "auth_ok":
				log.Debug().Caller().Msg("Registering app for push notifications.")
				request <- struct {
					Type           string `json:"type"`
					ID             int    `json:"id"`
					WebHookID      string `json:"webhook_id"`
					SupportConfirm bool   `json:"support_confirm"`
				}{
					Type:           "mobile_app/push_notification_channel",
					ID:             1,
					WebHookID:      config.WebhookID,
					SupportConfirm: false,
				}
			case "result":
				if !r.Success {
					log.Error().Msgf("Recieved error on websocket, %s: %s.", r.Error.Code, r.Error.Message)
					// reconnect <- true
				}
			case "event":
				agent.App.SendNotification(&fyne.Notification{
					Title:   r.Notification.Title,
					Content: r.Notification.Message,
				})
			default:
				log.Debug().Caller().Msgf("Received unhandled response %v", response)
			}
		}
	}
}
