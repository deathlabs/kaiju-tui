package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/deathlabs/kaiju-tui/internal/messages"
)

type KeycloakConfig struct {
	BaseURL  string
	Realm    string
	ClientID string
	Scopes   string
}

var authHTTPClient = &http.Client{Timeout: 10 * time.Second}

func (keycloakConfig KeycloakConfig) deviceEndpoint() string {
	return fmt.Sprintf(
		"%s/realms/%s/protocol/openid-connect/auth/device",
		keycloakConfig.BaseURL,
		keycloakConfig.Realm)
}

func (keycloakConfig KeycloakConfig) tokenEndpoint() string {
	return fmt.Sprintf(
		"%s/realms/%s/protocol/openid-connect/token",
		keycloakConfig.BaseURL,
		keycloakConfig.Realm,
	)
}

func RequestDeviceCodeCmd(cfg KeycloakConfig) tea.Cmd {
	return func() tea.Msg {
		var (
			cancel context.CancelFunc
			ctx    context.Context
			dcr    struct {
				DeviceCode              string `json:"device_code"`
				UserCode                string `json:"user_code"`
				VerificationURI         string `json:"verification_uri"`
				VerificationURIComplete string `json:"verification_uri_complete"`
				ExpiresIn               int    `json:"expires_in"`
				Interval                int    `json:"interval"`
			}
			err      error
			form     url.Values
			request  *http.Request
			response *http.Response
		)

		form = url.Values{}
		form.Set("client_id", cfg.ClientID)
		if cfg.Scopes != "" {
			form.Set("scope", cfg.Scopes)
		}

		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		request, err = http.NewRequestWithContext(ctx, http.MethodPost, cfg.deviceEndpoint(), strings.NewReader(form.Encode()))
		if err != nil {
			return messages.DeviceCodeMsg{Err: err}
		}
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		response, err = authHTTPClient.Do(request)
		if err != nil {
			return messages.DeviceCodeMsg{Err: err}
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			return messages.DeviceCodeMsg{Err: fmt.Errorf("device code request failed: %s", response.Status)}
		}

		if err = json.NewDecoder(response.Body).Decode(&dcr); err != nil {
			return messages.DeviceCodeMsg{Err: err}
		}

		return messages.DeviceCodeMsg{
			DeviceCode:              dcr.DeviceCode,
			UserCode:                dcr.UserCode,
			VerificationURI:         dcr.VerificationURI,
			VerificationURIComplete: dcr.VerificationURIComplete,
			ExpiresIn:               dcr.ExpiresIn,
			Interval:                dcr.Interval,
		}
	}
}

// PollTokenCmd polls Keycloak's token endpoint for the given device code.
func PollTokenCmd(cfg KeycloakConfig, deviceCode string) tea.Cmd {
	return func() tea.Msg {
		var (
			ctx      context.Context
			cancel   context.CancelFunc
			err      error
			form     url.Values
			request  *http.Request
			response *http.Response
			tr       struct {
				AccessToken  string `json:"access_token"`
				RefreshToken string `json:"refresh_token"`
				ExpiresIn    int    `json:"expires_in"`
				Error        string `json:"error"`
				ErrorDesc    string `json:"error_description"`
			}
		)

		form = url.Values{}
		form.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
		form.Set("device_code", deviceCode)
		form.Set("client_id", cfg.ClientID)

		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		request, err = http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			cfg.tokenEndpoint(),
			strings.NewReader(form.Encode()),
		)

		if err != nil {
			return messages.AuthResultMsg{Err: err, Done: true}
		}

		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		response, err = authHTTPClient.Do(request)
		if err != nil {
			return messages.AuthResultMsg{Err: err, Done: true}
		}
		defer response.Body.Close()

		if err = json.NewDecoder(response.Body).Decode(&tr); err != nil {
			return messages.AuthResultMsg{Err: err, Done: true}
		}

		switch {
		case tr.AccessToken != "":
			return messages.AuthResultMsg{
				AccessToken:  tr.AccessToken,
				RefreshToken: tr.RefreshToken,
				ExpiresIn:    tr.ExpiresIn,
				Done:         true,
			}
		case tr.Error == "authorization_pending", tr.Error == "slow_down":
			return messages.AuthResultMsg{}
		case tr.Error == "expired_token", tr.Error == "access_denied":
			return messages.AuthResultMsg{Err: fmt.Errorf("%s: %s", tr.Error, tr.ErrorDesc), Done: true}
		default:
			if tr.Error != "" {
				return messages.AuthResultMsg{
					Err: fmt.Errorf(
						"%s: %s",
						tr.Error,
						tr.ErrorDesc),
					Done: true}
			}
			return messages.AuthResultMsg{
				Err: fmt.Errorf(
					"unexpected token response: %s",
					response.Status),
				Done: true}
		}
	}
}

// AuthPollTickCmd schedules the next poll after the given interval.
func AuthPollTickCmd(intervalSeconds int) tea.Cmd {
	var duration time.Duration

	duration = time.Duration(intervalSeconds) * time.Second

	if duration <= 0 {
		duration = 5 * time.Second
	}

	return tea.Tick(duration, func(time.Time) tea.Msg {
		return messages.AuthPollTickMsg{}
	})
}
