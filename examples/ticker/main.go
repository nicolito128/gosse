package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/nicolito128/gosse"
)

var portFlag = flag.String("port", ":8080", "Network port")

func main() {
	flag.Parse()
	port := *portFlag

	eventCh := gosse.NewChannel()

	startTicker(eventCh)

	http.Handle("/", http.FileServer(http.Dir("static")))
	http.Handle("/events", eventCh)

	log.Printf("Server started on %s", port)
	log.Printf("Home: http://localhost%s/", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func startTicker(ch *gosse.Channel) {
	count := 0
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for range ticker.C {
			count++
			ch.Push(gosse.Event{
				ID:    fmt.Sprintf("%d", count),
				Event: "tick-update",
				Data:  fmt.Sprintf(`{"time_now": "%s"}`, time.Now().Format(time.RFC3339)),
			})
			log.Println("Sending a new tick...")
		}
	}()
}
