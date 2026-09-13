package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/deathlabs/kaiju-tui/internal/api/models"
	"github.com/deathlabs/kaiju-tui/internal/messages"
)

func CreateExercise(
	baseURL string,
	token string,
	exercise *models.ExerciseCreate,
) tea.Msg {
	var (
		body         []byte
		client       *http.Client
		err          error
		request      *http.Request
		response     *http.Response
		responseBody []byte
	)

	body, err = json.Marshal(exercise)
	if err != nil {
		return messages.ErrorMessage{Message: err}
	}

	request, err = http.NewRequest(
		http.MethodPost,
		strings.TrimRight(baseURL, "/")+"/api/v1/exercises/",
		bytes.NewReader(body),
	)
	if err != nil {
		return messages.ErrorMessage{Message: err}
	}

	request.Header.Set("Content-Type", "application/json")

	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}

	client = &http.Client{
		Timeout: 10 * time.Second,
	}

	response, err = client.Do(request)
	if err != nil {
		return messages.ErrorMessage{Message: err}
	}

	responseBody, err = io.ReadAll(response.Body)
	if err != nil {
		return messages.ErrorMessage{Message: err}
	}

	err = response.Body.Close()
	if err != nil {
		return messages.ErrorMessage{Message: err}
	}

	if response.StatusCode != http.StatusCreated {
		return messages.ErrorMessage{
			Message: fmt.Errorf(
				"create exercise returned %s: %s",
				response.Status,
				string(responseBody),
			),
		}
	}

	return messages.ExerciseCreatedMessage{
		StatusCode: response.StatusCode,
	}
}

func CreateExerciseCmd(
	baseURL string,
	token string,
	exercise *models.ExerciseCreate,
) tea.Cmd {
	return func() tea.Msg {
		return CreateExercise(baseURL, token, exercise)
	}
}
