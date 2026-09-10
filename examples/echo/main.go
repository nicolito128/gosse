package main

import (
	"fmt"
	"net/http"
	"sync"
	"uuid"

	"github.com/nicolito128/gosse"
)

var (
	mu      sync.RWMutex
	clients = map[string]*gosse.Channel{}
)

func main() {
	http.HandleFunc("/events", handleEvents)
	http.HandleFunc("/echo/{id}/{msg}", handleEcho)

	fmt.Println("Listening on :8080")
	http.ListenAndServe(":8080", nil)
}

func handleEvents(w http.ResponseWriter, r *http.Request) {
	c := gosse.Upgrade(w, r)
	id := uuid.New().String()

	mu.Lock()
	clients[id] = c
	mu.Unlock()

	defer func() {
		mu.Lock()
		delete(clients, id)
		mu.Unlock()
		c.Close()
	}()

	mid := gosse.NewMessage([]byte(id), "echo.id", 0)
	if _, err := c.Send(mid.Bytes()); err != nil {
		http.Error(w, "failed to send", http.StatusInternalServerError)
		return
	}

	<-c.Done()
}

func handleEcho(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	msg := r.PathValue("msg")

	mu.RLock()
	c, ok := clients[id]
	mu.RUnlock()

	if !ok {
		http.Error(w, "client not found", http.StatusNotFound)
		return
	}

	m := gosse.NewMessage([]byte(msg), "echo.msg", 0)
	if _, err := c.Send(m.Bytes()); err != nil {
		http.Error(w, "failed to send", http.StatusInternalServerError)
	}
}
