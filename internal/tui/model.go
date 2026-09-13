package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/deathlabs/kaiju-tui/internal/screens"
)

type Screen int

const (
	ScreenFacilitator Screen = iota
	ScreenParticipant
)

type Model struct {
	Screen      Screen
	Facilitator screens.FacilitatorScreen
	Participant screens.ParticipantScreen
}

func (model Model) Init() tea.Cmd {
	return nil
}

func (model Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "f":
			model.Screen = ScreenFacilitator
			return model, nil

		case "p":
			model.Screen = ScreenParticipant
			return model, nil
		}
	}

	switch model.Screen {
	case ScreenFacilitator:
		model.Facilitator, cmd = model.Facilitator.Update(msg)

	case ScreenParticipant:
		model.Participant, cmd = model.Participant.Update(msg)
	}

	return model, cmd
}

func (model Model) View() tea.View {
	switch model.Screen {
	case ScreenFacilitator:
		return model.Facilitator.View()

	case ScreenParticipant:
		return model.Participant.View()
	}

	return tea.NewView("unknown screen")
}
