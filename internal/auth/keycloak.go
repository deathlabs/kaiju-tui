package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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
			cancel             context.CancelFunc
			closeErr           error
			ctx                context.Context
			deviceCodeResponse DeviceCodeResponse
			err                error
			form               url.Values
			readErr            error
			request            *http.Request
			response           *http.Response
			responseBody       []byte
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

		responseBody, readErr = io.ReadAll(response.Body)
		closeErr = response.Body.Close()

		if readErr != nil {
			return messages.DeviceCodeMsg{Err: readErr}
		}

		if closeErr != nil {
			return messages.DeviceCodeMsg{Err: closeErr}
		}

		if response.StatusCode != http.StatusOK {
			return messages.DeviceCodeMsg{
				Err: fmt.Errorf(
					"device code request failed: %s",
					response.Status)}
		}

		err = json.Unmarshal(responseBody, &deviceCodeResponse)
		if err != nil {
			return messages.DeviceCodeMsg{Err: err}
		}

		return messages.DeviceCodeMsg{
			DeviceCode:              deviceCodeResponse.DeviceCode,
			UserCode:                deviceCodeResponse.UserCode,
			VerificationURI:         deviceCodeResponse.VerificationURI,
			VerificationURIComplete: deviceCodeResponse.VerificationURIComplete,
			ExpiresIn:               deviceCodeResponse.ExpiresIn,
			Interval:                deviceCodeResponse.Interval,
		}
	}
}

// PollTokenCmd polls Keycloak's token endpoint for the given device code.
func PollTokenCmd(cfg KeycloakConfig, deviceCode string) tea.Cmd {
	return func() tea.Msg {
		var (
			cancel        context.CancelFunc
			closeErr      error
			ctx           context.Context
			err           error
			form          url.Values
			readErr       error
			request       *http.Request
			response      *http.Response
			responseBody  []byte
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

		responseBody, readErr = io.ReadAll(response.Body)
		closeErr = response.Body.Close()

		if readErr != nil {
			return messages.AuthResultMsg{Err: readErr, Done: true}
		}

		if closeErr != nil {
			return messages.AuthResultMsg{Err: closeErr, Done: true}
		}

		err = json.Unmarshal(responseBody, &tokenResponse)
		if err != nil {
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
		case tokenResponse.Error == "authorization_pending",
			tokenResponse.Error == "slow_down":
			return messages.AuthResultMsg{}
		case tokenResponse.Error == "expired_token",
			tokenResponse.Error == "access_denied":
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

// AuthPollTickMsg creates a message to indicate the poll interval has elapsed.
func AuthPollTickMsg(time.Time) tea.Msg {
	return messages.AuthPollTickMsg{}
}

// AuthPollTickCmd schedules the next poll after the given interval.
func AuthPollTickCmd(intervalSeconds int) tea.Cmd {
	var duration time.Duration

	duration = time.Duration(intervalSeconds) * time.Second

	if duration <= 0 {
		duration = 5 * time.Second
	}

	return tea.Tick(duration, AuthPollTickMsg)
}
