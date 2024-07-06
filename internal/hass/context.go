// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package hass

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"

	"github.com/joshuar/go-hass-agent/internal/preferences"
)

type contextKey string

const (
	clientContextKey contextKey = "client"
)

func ContextSetClient(ctx context.Context, client *resty.Client) context.Context {
	newCtx := context.WithValue(ctx, clientContextKey, client)

	return newCtx
}

func ContextGetClient(ctx context.Context) *resty.Client {
	url, ok := ctx.Value(clientContextKey).(*resty.Client)
	if !ok {
		return nil
	}

	return url
}

func SetupContext(ctx context.Context) (context.Context, error) {
	prefs, err := preferences.ContextGetPrefs(ctx)
	if err != nil {
		return ctx, fmt.Errorf("could not setup hass context: %w", err)
	}

	ctx = ContextSetClient(ctx, NewDefaultHTTPClient(prefs.RestAPIURL))

	return ctx, nil
}
