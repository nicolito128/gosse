package gosse

import "fmt"

// Event ...
// https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events#fields
type Event struct {
	ID    string `json:"id"`
	Event string `json:"event"`
	Data  string `json:"data"`
	Retry int    `json:"retry"`
}

func (e Event) String() string {
	var s string
	if e.ID != "" {
		s += fmt.Sprintf("id: %s\n", e.ID)
	}

	if e.Event != "" {
		s += fmt.Sprintf("event: %s\n", e.Event)
	}

	if e.Retry > 0 {
		s += fmt.Sprintf("retry: %d\n", e.Retry)
	}

	s += fmt.Sprintf("data: %s\n\n", e.Data)

	return s
}
