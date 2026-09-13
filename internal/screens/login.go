package screens

import (
	"fmt"
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"github.com/deathlabs/kaiju-tui/internal/auth"
	"github.com/deathlabs/kaiju-tui/internal/messages"
)

type state int

const banner = `
 __  __     ______     __       __     __  __    
/\ \/ /    /\  __ \   /\ \     /\ \   /\ \/\ \   
\ \  _"-.  \ \  __ \  \ \ \   _\_\ \  \ \ \_\ \  
 \ \_\ \_\  \ \_\ \_\  \ \_\ /\_____\  \ \_____\ 
  \/_/\/_/   \/_/\/_/   \/_/ \/_____/   \/_____/ 
`

const (
	stateRequestingCode state = iota
	stateAwaitingUser
	stateAuthenticated
	stateError
)

type LoginScreen struct {
	AccessToken  string
	authConfig   auth.KeycloakConfig
	deadline     time.Time
	deviceCode   string
	Err          error
	interval     int
	RefreshToken string
	spinner      spinner.Model
	state        state
	userCode     string
	verifyURL    string
}

func NewLoginScreen(keycloakConfig auth.KeycloakConfig) LoginScreen {
	loginSpinner := spinner.New()
	loginSpinner.Spinner = spinner.Dot

	return LoginScreen{
		authConfig: keycloakConfig,
		spinner:    loginSpinner,
		state:      stateRequestingCode,
	}
}

func (loginScreen LoginScreen) Init() tea.Cmd {
	return tea.Batch(
		auth.RequestDeviceCodeCmd(loginScreen.authConfig),
		loginScreen.spinner.Tick,
	)
}

func (loginScreen LoginScreen) Authenticated() bool {
	return loginScreen.state == stateAuthenticated
}

func (loginScreen LoginScreen) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message := message.(type) {
	case tea.KeyPressMsg:
		switch message.String() {
		case "ctrl+c", "q":
			return loginScreen, tea.Quit
		}

	case spinner.TickMsg:
		if loginScreen.state != stateRequestingCode &&
			loginScreen.state != stateAwaitingUser {
			return loginScreen, nil
		}

		var cmd tea.Cmd
		loginScreen.spinner, cmd = loginScreen.spinner.Update(message)
		return loginScreen, cmd

	case messages.DeviceCodeMsg:
		if message.Err != nil {
			loginScreen.state = stateError
			loginScreen.Err = message.Err
			return loginScreen, nil
		}

		loginScreen.userCode = message.UserCode
		loginScreen.deviceCode = message.DeviceCode
		loginScreen.verifyURL = message.VerificationURI

		if message.VerificationURIComplete != "" {
			loginScreen.verifyURL = message.VerificationURIComplete
		}

		loginScreen.interval = message.Interval
		loginScreen.deadline = time.Now().Add(
			time.Duration(message.ExpiresIn) * time.Second,
		)

		loginScreen.state = stateAwaitingUser

		return loginScreen, auth.AuthPollTickCmd(loginScreen.interval)

	case messages.AuthPollTickMsg:
		if time.Now().After(loginScreen.deadline) {
			loginScreen.state = stateError
			loginScreen.Err = fmt.Errorf("device code expired, restart login")
			return loginScreen, nil
		}

		return loginScreen, auth.PollTokenCmd(
			loginScreen.authConfig,
			loginScreen.deviceCode,
		)

	case messages.AuthResultMsg:
		if message.Err != nil {
			loginScreen.state = stateError
			loginScreen.Err = message.Err
			return loginScreen, nil
		}

		if message.AccessToken != "" {
			loginScreen.state = stateAuthenticated
			loginScreen.AccessToken = message.AccessToken
			loginScreen.RefreshToken = message.RefreshToken
			return loginScreen, nil
		}

		return loginScreen, auth.AuthPollTickCmd(loginScreen.interval)
	}

	return loginScreen, nil
}

func (loginScreen LoginScreen) View() tea.View {
	var (
		screen string
		view   tea.View
	)

	switch loginScreen.state {
	case stateRequestingCode:
		screen = fmt.Sprintf(
			"%s Requesting device code...\n",
			loginScreen.spinner.View(),
		)

	case stateAwaitingUser:
		screen = fmt.Sprintf(
			"To sign in, open:\n\n  %s\n\nWaiting for authentication...\n\nPress q to quit.\n",
			loginScreen.verifyURL,
		)

	case stateAuthenticated:
		screen = "Authenticated.\n"

	case stateError:
		screen = fmt.Sprintf(
			"Authentication failed: %v\n\nPress q to quit.\n",
			loginScreen.Err,
		)
	}

	view = tea.NewView(banner + "\n" + screen)
	view.AltScreen = true
	return view
}
