package screens

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/deathlabs/kaiju-tui/internal/api"
	"github.com/deathlabs/kaiju-tui/internal/api/models"
	"github.com/deathlabs/kaiju-tui/internal/messages"
)

type ExerciseFormData struct {
	Start string
	End   string
}

type ExerciseScreen struct {
	form *huh.Form

	exercise *models.ExerciseCreate
	data     *ExerciseFormData

	baseURL   string
	token     string
	submitted bool
	created   bool
	err       error
}

func (screen *ExerciseScreen) SetToken(token string) {
	screen.token = token
}

func NewExerciseScreen(url string) ExerciseScreen {
	var (
		exercise *models.ExerciseCreate
		data     *ExerciseFormData
	)

	exercise = &models.ExerciseCreate{}
	data = &ExerciseFormData{
		Start: "2026-09-20T13:00:00-04:00",
		End:   "2026-09-20T15:00:00-04:00",
	}

	return ExerciseScreen{
		exercise: exercise,
		data:     data,
		baseURL:  url,
		token:    "",

		form: huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[models.ExerciseType]().
					Title("TTX Type").
					Options(
						huh.NewOption(
							"Discussion",
							models.ExerciseTypeDiscussion,
						),
						huh.NewOption(
							"Discussion and Hands-On",
							models.ExerciseTypeDiscussionHandsOn,
						),
					).
					Value(&exercise.Type),

				huh.NewInput().
					Title("Title").
					Value(&exercise.Title),

				huh.NewText().
					Title("Scenario").
					Value(&exercise.Scenario),

				huh.NewInput().
					Title("Scheduled Start Time").
					Placeholder("2026-09-20T13:00:00-04:00").
					Value(&data.Start),

				huh.NewInput().
					Title("Scheduled End Time").
					Placeholder("2026-09-20T15:00:00-04:00").
					Value(&data.End),
			),
		),
	}
}

func (screen ExerciseScreen) Init() tea.Cmd {
	if screen.form == nil {
		return nil
	}

	return screen.form.Init()
}

func (screen ExerciseScreen) Update(msg tea.Msg) (ExerciseScreen, tea.Cmd) {
	var (
		cmd   tea.Cmd
		err   error
		form  *huh.Form
		model huh.Model
		ok    bool
		start time.Time
		end   time.Time
	)

	switch msg := msg.(type) {
	case messages.ExerciseCreatedMessage:
		screen.created = true
		return screen, nil

	case messages.ErrorMessage:
		screen.err = msg.Message
		return screen, nil
	}

	if screen.form == nil {
		return screen, nil
	}

	model, cmd = screen.form.Update(msg)

	form, ok = model.(*huh.Form)
	if ok {
		screen.form = form
	}

	if screen.form.State == huh.StateCompleted && !screen.submitted {
		start, err = time.Parse(time.RFC3339, screen.data.Start)
		if err != nil {
			screen.err = err
			return screen, nil
		}

		end, err = time.Parse(time.RFC3339, screen.data.End)
		if err != nil {
			screen.err = err
			return screen, nil
		}

		screen.exercise.ScheduledStartTime = start
		screen.exercise.ScheduledEndTime = end

		screen.submitted = true

		return screen, tea.Batch(
			cmd,
			api.CreateExerciseCmd(
				screen.baseURL,
				screen.token,
				screen.exercise,
			),
		)
	}

	return screen, cmd
}

func (screen ExerciseScreen) View() tea.View {
	if screen.form == nil {
		return tea.NewView("exercise form not initialized")
	}

	if screen.err != nil {
		return tea.NewView(
			fmt.Sprintf("Failed to create exercise:\n%s", screen.err),
		)
	}

	if screen.created {
		return tea.NewView("Exercise created successfully.")
	}

	if screen.submitted {
		return tea.NewView("Creating exercise...")
	}

	return tea.NewView(screen.form.View())
}
