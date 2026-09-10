package client

import (
	"net/http"
	"time"

	tea "charm.land/bubbletea/v2"
)

type errMsg struct{ error }

type statusMsg int

func checkServer(url string) tea.Msg {
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
		return errMsg{err}
	}
	defer response.Body.Close() // nolint:errcheck

	return statusMsg(response.StatusCode)
}

func checkServerCmd(url string) tea.Cmd {
	return func() tea.Msg {
		return checkServer(url)
	}
}
