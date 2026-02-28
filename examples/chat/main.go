package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"

	"github.com/nicolito128/gosse"
)

var portFlag = flag.String("port", ":8080", "Network port")

type Chatroom struct {
	*gosse.Channel
	messages []Message
}

func NewChatroom() *Chatroom {
	return &Chatroom{
		Channel:  gosse.NewChannel(),
		messages: make([]Message, 0),
	}
}

type Message struct {
	Author  string `json:"author"`
	Content string `json:"content"`
}

func main() {
	flag.Parse()
	port := *portFlag

	globalChat := NewChatroom()

	http.Handle("/", http.FileServer(http.Dir("static")))
	http.Handle("/join-chatroom", globalChat)
	http.HandleFunc("POST /publish", HandleNewMessage(globalChat))

	log.Printf("Chatroom at http://localhost%s/ - Press CTRL+C to exit", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func HandleNewMessage(cr *Chatroom) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var msg Message

		err := json.NewDecoder(r.Body).Decode(&msg)
		if err != nil {
			http.Error(w, "error decoding the message", http.StatusBadRequest)
			return
		}

		data, err := json.Marshal(msg)
		if err != nil {
			http.Error(w, "error encoding the response", http.StatusInternalServerError)
			return
		}

		newMsg := gosse.Event{
			Event: "new-chatroom-message",
			Data:  string(data),
		}

		cr.Push(newMsg)
	}
}
