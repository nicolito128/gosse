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
	http.HandleFunc("/chat/{id}/{msg}", handleChat)

	fmt.Println("Chat server listening on :8080")
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

	welcome := gosse.NewMessage("chat.id", []byte(id), -1)
	c.Send(welcome.Bytes())

	broadcast(gosse.NewMessage("chat.join", []byte(id+" joined"), -1))

	<-c.Done()

	broadcast(gosse.NewMessage("chat.leave", []byte(id+" left"), -1))
}

func handleChat(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	msg := r.PathValue("msg")

	mu.RLock()
	_, ok := clients[id]
	mu.RUnlock()

	if !ok {
		http.Error(w, "unknown client id", http.StatusNotFound)
		return
	}

	text := fmt.Sprintf("%s: %s", id, msg)
	broadcast(gosse.NewMessage("chat.msg", []byte(text), -1))
}

func broadcast(m *gosse.Message) {
	mu.RLock()
	defer mu.RUnlock()

	payload := []byte(m.String())
	for _, c := range clients {
		c.Send(payload)
	}
}
