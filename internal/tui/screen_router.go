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

type ScreenRouter struct {
	Facilitator   screens.FacilitatorScreen
	Participant   screens.ParticipantScreen
	Exercise      screens.ExerciseScreen
	CurrentScreen Screen
	Token         string
}

func NewScreenRouter(token string) ScreenRouter {
	return ScreenRouter{
		Facilitator:   screens.NewFacilitatorScreen(),
		Participant:   screens.NewParticipantScreen(),
		Exercise:      screens.NewExerciseScreen("https://kaiju.uds.dev"),
		CurrentScreen: ScreenFacilitator,
		Token:         token,
	}
}

func (router ScreenRouter) Init() tea.Cmd {
	return nil
}

func (router ScreenRouter) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return router, tea.Quit
		case "e":
			router.Exercise.SetToken(router.Token)
			router.CurrentScreen = ScreenExercise
			return router, router.Exercise.Init()
		case "ctrl+f":
			router.CurrentScreen = ScreenFacilitator
			return router, nil
		case "ctrl+p":
			router.CurrentScreen = ScreenParticipant
			return router, nil
		}
	}

	switch router.CurrentScreen {
	case ScreenFacilitator:
		router.Facilitator, cmd = router.Facilitator.Update(msg)
	case ScreenParticipant:
		router.Participant, cmd = router.Participant.Update(msg)
	case ScreenExercise:
		router.Exercise, cmd = router.Exercise.Update(msg)
	}

	return router, cmd
}

func (router ScreenRouter) View() tea.View {
	var view tea.View

	switch router.CurrentScreen {
	case ScreenFacilitator:
		view = router.Facilitator.View()
	case ScreenParticipant:
		view = router.Participant.View()
	case ScreenExercise:
		view = router.Exercise.View()
	default:
		view = tea.NewView("unknown screen")
	}

	view.AltScreen = true
	return view
}
