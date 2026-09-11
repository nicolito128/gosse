package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	addr  = flag.String("addr", "http://localhost:8080", "server base address")
	retry = flag.Duration("retry", 3*time.Second, "reconnect retry duration")
)

func main() {
	flag.Parse()
	base := strings.TrimRight(*addr, "/")

	lastEventID := ""

	for {
		newID, newRetry, err := connectAndListen(base, lastEventID)

		if newID != "" {
			lastEventID = newID
		}
		if newRetry > 0 {
			*retry = time.Duration(newRetry) * time.Millisecond
		}

		if err != nil {
			log.Printf("stream ended with error: %v", err)
		} else {
			log.Println("stream ended")
		}

		log.Printf("reconnecting in %s (Last-Event-ID: %q)", *retry, lastEventID)
		time.Sleep(*retry)
	}
}

// connectAndListen opens one connection to /events, sending lastEventID
// (if any) via the Last-Event-ID header, and reads events until the
// stream ends. It returns the most recent event id seen and the most
// recent retry value announced by the server, so the caller can carry
// both into the next reconnect attempt.
func connectAndListen(base, lastEventID string) (newLastEventID string, retryMillis int, err error) {
	req, err := http.NewRequest(http.MethodGet, base+"/events", nil)
	if err != nil {
		return lastEventID, 0, err
	}
	if lastEventID != "" {
		req.Header.Set("Last-Event-ID", lastEventID)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return lastEventID, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return lastEventID, 0, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	scanner := bufio.NewScanner(resp.Body)

	var event, data, id string
	newLastEventID = lastEventID

	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case line == "":
			if event == "" && data == "" {
				return
			}
			fmt.Printf("[%s] %s (id=%s)\n", event, data, id)
			if id != "" {
				newLastEventID = id
			}
			event, data = "", ""

		case strings.HasPrefix(line, "event: "):
			event = strings.TrimPrefix(line, "event: ")

		case strings.HasPrefix(line, "id: "):
			id = strings.TrimPrefix(line, "id: ")

		case strings.HasPrefix(line, "retry: "):
			if n, err := strconv.Atoi(strings.TrimPrefix(line, "retry: ")); err == nil {
				retryMillis = n
			}

		case strings.HasPrefix(line, "data: "):
			if data != "" {
				data += "\n"
			}
			data += strings.TrimPrefix(line, "data: ")
		}
	}

	return newLastEventID, retryMillis, scanner.Err()
}
