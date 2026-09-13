package screens

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

type ExerciseScreen struct {
	form *huh.Form
}

func NewExerciseScreen() ExerciseScreen {
	return ExerciseScreen{
		form: huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("TTX Type").
					Key("type").
					Options(huh.NewOptions(
						"Discussion",
						"Discussion and Hands-On")...),
				huh.NewSelect[int]().
					Title("Choose your level").
					Key("level").
					Options(huh.NewOptions(1, 20, 9999)...),
			),
		),
	}
}

func (m ExerciseScreen) Init() tea.Cmd {
	if m.form == nil {
		return nil
	}

	return m.form.Init()
}

func (screen ExerciseScreen) Update(msg tea.Msg) (ExerciseScreen, tea.Cmd) {
	var (
		cmd   tea.Cmd
		form  *huh.Form
		model huh.Model
		ok    bool
	)

	if screen.form == nil {
		return screen, nil
	}

	model, cmd = screen.form.Update(msg)
	form, ok = model.(*huh.Form)
	if ok {
		screen.form = form
	}

	return screen, cmd
}

func (screen ExerciseScreen) View() tea.View {
	if screen.form == nil {
		return tea.NewView("exercise form not initialized")
	}

	if screen.form.State == huh.StateCompleted {
		class := screen.form.GetString("type")
		level := screen.form.GetInt("level")

		return tea.NewView(
			fmt.Sprintf("You selected: %s, Lvl. %d", class, level),
		)
	}

	return tea.NewView(screen.form.View())
}
