package main

import (
	"flag"
	"fmt"
	"net/http"
	"sync"
	"uuid"

	"github.com/nicolito128/gosse"
)

var (
	addr = flag.String("addr", ":8080", "server base address")
)

var (
	mu      sync.RWMutex
	clients = map[string]*gosse.Channel{}
)

func main() {
	flag.Parse()

	http.HandleFunc("/events", handleEvents)
	http.HandleFunc("/echo/{id}/{msg}", handleEcho)

	fmt.Println("Listening on", *addr)
	http.ListenAndServe(*addr, nil)
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

	mid := gosse.NewMessage("echo.id", []byte(id), -1)
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

	m := gosse.NewMessage("echo.msg", []byte(msg), -1)
	if _, err := c.Send(m.Bytes()); err != nil {
		http.Error(w, "failed to send", http.StatusInternalServerError)
	}
}
