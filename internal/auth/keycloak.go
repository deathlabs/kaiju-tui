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

type DeviceCodeResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

type TokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int    `json:"expires_in"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

var httpClient = &http.Client{Timeout: 10 * time.Second}

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
			cancel   context.CancelFunc
			ctx      context.Context
			dcr      DeviceCodeResponse
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

		response, err = httpClient.Do(request)
		if err != nil {
			return messages.DeviceCodeMsg{Err: err}
		}
		defer func() error {
			err = response.Body.Close()
			if err != nil {
				return err
			}
			return nil
		}()

		if response.StatusCode != http.StatusOK {
			return messages.DeviceCodeMsg{Err: fmt.Errorf("device code request failed: %s", response.Status)}
		}

		err = json.NewDecoder(response.Body).Decode(&dcr)
		if err != nil {
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
			ctx           context.Context
			cancel        context.CancelFunc
			err           error
			form          url.Values
			request       *http.Request
			response      *http.Response
			tokenResponse TokenResponse
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

		response, err = httpClient.Do(request)
		if err != nil {
			return messages.AuthResultMsg{Err: err, Done: true}
		}
		defer func() error {
			err = response.Body.Close()
			if err != nil {
				return err
			}
			return nil
		}()

		if err = json.NewDecoder(response.Body).Decode(&tokenResponse); err != nil {
			return messages.AuthResultMsg{Err: err, Done: true}
		}

		switch {
		case tokenResponse.AccessToken != "":
			return messages.AuthResultMsg{
				AccessToken:  tokenResponse.AccessToken,
				RefreshToken: tokenResponse.RefreshToken,
				ExpiresIn:    tokenResponse.ExpiresIn,
				Done:         true,
			}
		case tokenResponse.Error == "authorization_pending", tokenResponse.Error == "slow_down":
			return messages.AuthResultMsg{}
		case tokenResponse.Error == "expired_token", tokenResponse.Error == "access_denied":
			return messages.AuthResultMsg{
				Err: fmt.Errorf(
					"%s: %s",
					tokenResponse.Error,
					tokenResponse.ErrorDescription),
				Done: true}
		default:
			if tokenResponse.Error != "" {
				return messages.AuthResultMsg{
					Err: fmt.Errorf(
						"%s: %s",
						tokenResponse.Error,
						tokenResponse.ErrorDescription),
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
