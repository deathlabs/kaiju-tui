package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/deathlabs/kaiju-tui/internal/auth"
	"github.com/deathlabs/kaiju-tui/internal/screens"
)

// App wires the screens together.
type App struct {
	auth  screens.LoginScreen
	model Model
	ready bool
}

func NewApp(authConfig auth.KeycloakConfig, model Model) App {
	return App{
		auth:  screens.NewLoginScreen(authConfig),
		model: model,
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
		app.auth = model.(screens.LoginScreen)

		if app.auth.Authenticated() {
			app.ready = true
			app.model.Facilitator.Token = app.auth.AccessToken
			return app, app.model.Init()
		}

		return app, cmd
	}

	model, cmd = app.model.Update(msg)
	app.model = model.(Model)

	return app, cmd
}

func (app App) View() tea.View {
	if !app.ready {
		return app.auth.View()
	}

	return app.model.View()
}
