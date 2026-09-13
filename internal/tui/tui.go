package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/deathlabs/kaiju-tui/internal/auth"
	"github.com/deathlabs/kaiju-tui/internal/screens"
)

type TUI struct {
	LoginScreen  screens.LoginScreen
	ScreenRouter ScreenRouter
	ready        bool
}

func NewTUI(authConfig auth.KeycloakConfig, screenRouter ScreenRouter) TUI {
	return TUI{
		LoginScreen:  screens.NewLoginScreen(authConfig),
		ScreenRouter: screenRouter,
	}
}

func (tui TUI) Init() tea.Cmd {
	return tui.LoginScreen.Init()
}

func (tui TUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd          tea.Cmd
		screenRouter tea.Model
	)

	if !tui.ready {
		screenRouter, cmd = tui.LoginScreen.Update(msg)
		tui.LoginScreen = screenRouter.(screens.LoginScreen)

		if tui.LoginScreen.Authenticated() {
			tui.ready = true
			tui.ScreenRouter.Token = tui.LoginScreen.AccessToken
			tui.ScreenRouter.Facilitator.Token = tui.LoginScreen.AccessToken
			return tui, tui.ScreenRouter.Init()
		}

		return tui, cmd
	}

	screenRouter, cmd = tui.ScreenRouter.Update(msg)
	tui.ScreenRouter = screenRouter.(ScreenRouter)

	return tui, cmd
}

func (tui TUI) View() tea.View {
	if !tui.ready {
		return tui.LoginScreen.View()
	}

	return tui.ScreenRouter.View()
}
