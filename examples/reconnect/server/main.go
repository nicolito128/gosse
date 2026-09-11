package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/nicolito128/gosse"
)

func main() {
	http.HandleFunc("/events", handleEvents)

	fmt.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// handleEvents streams an incrementing counter to the client, one tick
// per second. Every 5 ticks it deliberately drops the connection, to
// simulate a flaky network and force the client to reconnect.
//
// A spec-compliant SSE client sends back the last event ID it received
// in the "Last-Event-ID" header on reconnect, so the server can resume
// the stream instead of starting over from zero.
func handleEvents(w http.ResponseWriter, r *http.Request) {
	c := gosse.Upgrade(w, r)
	defer c.Close()

	start := 0
	if last := r.Header.Get("Last-Event-ID"); last != "" {
		if n, err := strconv.Atoi(last); err == nil {
			start = n + 1
			log.Printf("client resuming from event id %d", start)
		}
	} else {
		log.Println("client connected fresh, starting from 0")
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	n := start
	sent := 0

	for {
		select {
		case <-c.Done():
			log.Println("client disconnected:", c.Error())
			return

		case <-ticker.C:
			msg := gosse.NewMessage(
				"counter.tick",
				[]byte(fmt.Sprintf("tick #%d", n)),
				gosse.UnspecifiedRetry,
			)
			msg.ID = strconv.Itoa(n)

			if _, err := c.Send([]byte(msg.String())); err != nil {
				return
			}

			n++
			sent++

			// Simulate a flaky connection: drop it every 5 messages.
			if sent%5 == 0 {
				log.Println("simulating a dropped connection")
				return
			}
		}
	}
}
