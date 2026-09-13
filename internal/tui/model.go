package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/deathlabs/kaiju-tui/internal/screens"
)

type Screen int

const (
	ScreenFacilitator Screen = iota
	ScreenParticipant
	ScreenExercise
)

type Model struct {
	Screen      Screen
	Facilitator screens.FacilitatorScreen
	Participant screens.ParticipantScreen
	Exercise    screens.ExerciseScreen
}

func (model Model) Init() tea.Cmd {
	return nil
}

func (model Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return model, tea.Quit
		case "e":
			model.Screen = ScreenExercise
			return model, model.Exercise.Init()
		case "ctrl+f":
			model.Screen = ScreenFacilitator
			return model, nil
		case "ctrl+p":
			model.Screen = ScreenParticipant
			return model, nil
		}
	}

	switch model.Screen {
	case ScreenFacilitator:
		model.Facilitator, cmd = model.Facilitator.Update(msg)

	case ScreenParticipant:
		model.Participant, cmd = model.Participant.Update(msg)

	case ScreenExercise:
		model.Exercise, cmd = model.Exercise.Update(msg)
	}

	return model, cmd
}

func (model Model) View() tea.View {
	switch model.Screen {
	case ScreenFacilitator:
		return model.Facilitator.View()

	case ScreenParticipant:
		return model.Participant.View()

	case ScreenExercise:
		return model.Exercise.View()
	}

	return tea.NewView("unknown screen")
}
