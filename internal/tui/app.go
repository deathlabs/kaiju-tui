package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/deathlabs/kaiju-tui/internal/auth"
)

// App wraps the auth screen and main screen.
type App struct {
	auth   auth.DeviceFlowLoginScreen
	server Model
	ready  bool
}

func NewApp(cfg auth.KeycloakConfig, server Model) App {
	return App{
		auth:   auth.New(cfg),
		server: server,
	}
}

func (app App) Init() tea.Cmd {
	return app.auth.Init()
}

func (app App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd   tea.Cmd
		model tea.Model
	)

	if !app.ready {
		model, cmd = app.auth.Update(msg)
		app.auth = model.(auth.DeviceFlowLoginScreen)

		if app.auth.Authenticated() {
			app.ready = true
			app.server.Token = app.auth.AccessToken
			return app, app.server.Init()
		}

		return app, cmd
	}

	model, cmd = app.server.Update(msg)
	app.server = model.(Model)
	return app, cmd
}

func (app App) View() tea.View {
	if !app.ready {
		return app.auth.View()
	}
	return app.server.View()
}
