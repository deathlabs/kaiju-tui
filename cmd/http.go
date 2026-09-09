package cmd

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
