// client.go
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var (
	addr = flag.String("addr", "http://localhost:8080", "server base address")
)

func main() {
	flag.Parse()

	base := strings.TrimRight(*addr, "/")

	clientID := make(chan string, 1)

	go listen(base, clientID)

	id := <-clientID
	fmt.Printf("connected as %s\n", id)
	fmt.Println("type a message and press enter to send it, ctrl+c to quit")

	send(base, id)
}

func listen(base string, clientID chan<- string) {
	resp, err := http.Get(base + "/events")
	if err != nil {
		log.Fatalf("connecting to %s/events: %v", base, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("unexpected status: %s", resp.Status)
	}

	scanner := bufio.NewScanner(resp.Body)

	var event, data string
	first := true

	flush := func() {
		if event == "" && data == "" {
			return
		}

		if event == "chat.id" && first {
			clientID <- data
			first = false
		} else {
			fmt.Printf("[%s] %s\n", event, data)
		}

		event, data = "", ""
	}

	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case line == "":
			flush()

		case strings.HasPrefix(line, "event: "):
			event = strings.TrimPrefix(line, "event: ")

		case strings.HasPrefix(line, "data: "):
			if data != "" {
				data += "\n"
			}
			data += strings.TrimPrefix(line, "data: ")
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("reading event stream: %v", err)
	}

	log.Println("connection closed by server")
	os.Exit(0)
}

func send(base, id string) {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		msg := strings.TrimSpace(scanner.Text())
		if msg == "" {
			continue
		}

		u := fmt.Sprintf("%s/chat/%s/%s", base, url.PathEscape(id), url.PathEscape(msg))

		resp, err := http.Get(u)
		if err != nil {
			log.Printf("send failed: %v", err)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("server returned %s", resp.Status)
		}
	}
}
