package tui

import (
	"net/http"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/deathlabs/kaiju-tui/internal/messages"
)

func CheckServer(url string, token string) tea.Msg {
	var (
		client   *http.Client
		err      error
		request  *http.Request
		response *http.Response
		timeout  time.Duration
	)

	timeout = 10 * time.Second

	client = &http.Client{
		Timeout: timeout,
	}

	request, err = http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return messages.ErrorMessage{Message: err}
	}

	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}

	response, err = client.Do(request)
	if err != nil {
		return messages.ErrorMessage{Message: err}
	}
	defer func() error {
		err = response.Body.Close()
		if err != nil {
			return err
		}
		return nil
	}()

	return messages.StatusMessage(response.StatusCode)
}

func CheckServerCmd(url string, token string) tea.Cmd {
	return func() tea.Msg {
		return CheckServer(url, token)
	}
}

func TickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return messages.TickMessage(t)
	})
}
