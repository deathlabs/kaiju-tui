package screens

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type ParticipantScreen struct {
	Content  string
	Ready    bool
	Viewport viewport.Model
	Token    string
}

func (screen ParticipantScreen) Init() tea.Cmd {
	return nil
}

func (screen ParticipantScreen) Update(msg tea.Msg) (ParticipantScreen, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return screen, tea.Quit
		}

	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(screen.headerView())
		footerHeight := lipgloss.Height(screen.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		if !screen.Ready {
			screen.Viewport = viewport.New(
				viewport.WithWidth(msg.Width),
				viewport.WithHeight(msg.Height-verticalMarginHeight),
			)

			screen.Viewport.YPosition = headerHeight
			screen.Viewport.SetContent(screen.Content)

			screen.Ready = true
		} else {
			screen.Viewport.SetWidth(msg.Width)
			screen.Viewport.SetHeight(
				msg.Height - verticalMarginHeight,
			)
		}
	}

	if screen.Ready {
		screen.Viewport, cmd = screen.Viewport.Update(msg)
	}

	return screen, cmd
}

func (screen ParticipantScreen) View() tea.View {
	var view tea.View

	view.AltScreen = true
	view.MouseMode = tea.MouseModeCellMotion

	if !screen.Ready {
		view.SetContent("\n  Initializing...")
		return view
	}

	view.SetContent(
		fmt.Sprintf(
			"%s\n%s\n%s",
			screen.headerView(),
			screen.Viewport.View(),
			screen.footerView(),
		),
	)

	return view
}

func (screen ParticipantScreen) headerView() string {
	title := "Participant Screen"

	line := strings.Repeat(
		"─",
		max(0, screen.Viewport.Width()-lipgloss.Width(title)),
	)

	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		title,
		line,
	)
}

func (screen ParticipantScreen) footerView() string {
	info := fmt.Sprintf(
		" %3.f%% ",
		screen.Viewport.ScrollPercent()*100,
	)

	line := strings.Repeat(
		"─",
		max(0, screen.Viewport.Width()-lipgloss.Width(info)),
	)

	status := lipgloss.JoinHorizontal(
		lipgloss.Center,
		line,
		info,
	)

	controls := "\n↑/↓: Scroll | Facilitator: f | Participant: p | Quit: q"

	return status + controls
}
