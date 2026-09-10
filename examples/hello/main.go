package main

import (
	"net/http"

	"github.com/nicolito128/gosse"
)

func main() {
	http.HandleFunc("/events", handleEvents)
	http.ListenAndServe(":8080", nil)
}

func handleEvents(w http.ResponseWriter, r *http.Request) {
	c := gosse.Upgrade(w, r)
	defer c.Close()

	msg := gosse.NewMessage([]byte("hello, goose!"), "greeting", 0)
	c.Send(msg.Bytes())

	<-c.Done()
}
