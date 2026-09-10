package tools

import (
	"net/http"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/deathlabs/kaiju-tui/internal/messages"
)

func CheckServer(url string) tea.Msg {
	var (
		client   *http.Client
		err      error
		response *http.Response
	)

	// This is only to test the progress bar bubble.
	time.Sleep(5 * time.Second)

	client = &http.Client{
		Timeout: 10 * time.Second,
	}

	response, err = client.Get(url)
	if err != nil {
		return messages.ErrorMessage{Message: err}
	}
	defer response.Body.Close()

	return messages.StatusMessage(response.StatusCode)
}

func CheckServerCmd(url string) tea.Cmd {
	return func() tea.Msg {
		return CheckServer(url)
	}
}

func TickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return messages.TickMessage(t)
	})
}
