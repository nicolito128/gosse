package gosse

import (
	"bytes"
	"encoding/json/v2"
	"fmt"
	"strconv"
)

const (
	// UnspecifiedRetry indicates that the retry value is not specified and should use the default.
	UnspecifiedRetry = -1
	// Default retry reconnection time in milliseconds
	DefaultRetry = 3000
)

// Message represents a single Server-Sent Event.
type Message struct {
	Event string
	Data  []byte
	ID    string
	Retry int
}

// NewMessage creates a new Message. If retry < 0, DefaultRetry is used.
func NewMessage(event string, data []byte, retry int) *Message {
	if retry < 0 {
		retry = DefaultRetry
	}
	return &Message{
		Data:  data,
		Event: event,
		Retry: retry,
	}
}

// String formats the Message according to the SSE wire format:
//
//	event: <event>\n
//	id: <id>\n
//	retry: <retry>\n
//	data: <line>\n   (one per line in Data)
//	\n
func (m *Message) String() string {
	var buf bytes.Buffer

	if m.Event != "" {
		fmt.Fprintf(&buf, "event: %s\n", m.Event)
	}

	if m.ID != "" {
		fmt.Fprintf(&buf, "id: %s\n", m.ID)
	}

	if m.Retry >= 0 {
		buf.WriteString("retry: ")
		buf.WriteString(strconv.Itoa(m.Retry))
		buf.WriteByte('\n')
	}

	for _, line := range bytes.Split(m.Data, []byte("\n")) {
		buf.WriteString("data: ")
		buf.Write(line)
		buf.WriteByte('\n')
	}

	buf.WriteByte('\n')

	return buf.String()
}

// Bytes returns the Message as a byte slice.
func (m *Message) Bytes() []byte {
	return []byte(m.String())
}

func (m *Message) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"event": "%s", "id": "%s", "retry": %d, "data": "%s"}`, m.Event, m.ID, m.Retry, string(m.Data))), nil
}

func (m *Message) UnmarshalJSON(data []byte, opts ...json.Options) error {
	var err error
	var event, id string
	var retry int
	var dataBytes []byte

	if err = json.Unmarshal(data, &event, opts...); err != nil {
		return err
	}
	m.Event = event

	if err = json.Unmarshal(data, &id, opts...); err != nil {
		return err
	}
	m.ID = id

	if err = json.Unmarshal(data, &retry, opts...); err != nil {
		return err
	}
	m.Retry = retry

	if err = json.Unmarshal(data, &dataBytes, opts...); err != nil {
		return err
	}
	m.Data = dataBytes

	return nil
}
