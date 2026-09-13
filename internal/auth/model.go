package auth

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/deathlabs/kaiju-tui/internal/messages"
)

type state int

const (
	stateRequestingCode state = iota
	stateAwaitingUser
	stateAuthenticated
	stateError
)

type DeviceFlowLoginScreen struct {
	cfg          KeycloakConfig
	state        state
	deviceCode   string
	userCode     string
	verifURL     string
	interval     int
	deadline     time.Time
	AccessToken  string
	RefreshToken string
	Err          error
}

func New(keycloakConfig KeycloakConfig) DeviceFlowLoginScreen {
	return DeviceFlowLoginScreen{cfg: keycloakConfig, state: stateRequestingCode}
}

func (model DeviceFlowLoginScreen) Init() tea.Cmd {
	return RequestDeviceCodeCmd(model.cfg)
}

func (model DeviceFlowLoginScreen) Authenticated() bool {
	return model.state == stateAuthenticated
}

func (model DeviceFlowLoginScreen) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message := message.(type) {
	case tea.KeyPressMsg:
		switch message.String() {
		case "ctrl+c", "q":
			return model, tea.Quit
		}

	case messages.DeviceCodeMsg:
		if message.Err != nil {
			model.state = stateError
			model.Err = message.Err
			return model, nil
		}
		model.userCode = message.UserCode
		model.deviceCode = message.DeviceCode
		model.verifURL = message.VerificationURI
		if message.VerificationURIComplete != "" {
			model.verifURL = message.VerificationURIComplete
		}
		model.interval = message.Interval
		model.deadline = time.Now().Add(time.Duration(message.ExpiresIn) * time.Second)
		model.state = stateAwaitingUser
		return model, AuthPollTickCmd(model.interval)

	case messages.AuthPollTickMsg:
		if time.Now().After(model.deadline) {
			model.state = stateError
			model.Err = fmt.Errorf("device code expired, restart login")
			return model, nil
		}
		return model, PollTokenCmd(model.cfg, model.deviceCode)

	case messages.AuthResultMsg:
		if message.Err != nil {
			model.state = stateError
			model.Err = message.Err
			return model, nil
		}
		if message.AccessToken != "" {
			model.state = stateAuthenticated
			model.AccessToken = message.AccessToken
			model.RefreshToken = message.RefreshToken
			return model, nil
		}
		return model, AuthPollTickCmd(model.interval)
	}

	return model, nil
}

func (model DeviceFlowLoginScreen) View() tea.View {
	var ui string

	switch model.state {
	case stateRequestingCode:
		ui = "Requesting device code...\n"
	case stateAwaitingUser:
		ui = fmt.Sprintf(
			"To sign in, open:\n\n  %s\n\nWaiting for authentication...\n\nPress q to quit.\n",
			model.verifURL,
		)
	case stateAuthenticated:
		ui = "Authenticated.\n"
	case stateError:
		ui = fmt.Sprintf("Auth failed: %v\n\nPress q to quit.\n", model.Err)
	}

	return tea.NewView(ui)
}
