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

var clientID string

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
	colorBold   = "\033[1m"
)

func main() {
	flag.Parse()

	base := strings.TrimRight(*addr, "/")

	go listen(base)

	send(base)
}

func listen(base string) {
	resp, err := http.Get(base + "/events")
	if err != nil {
		logErrorf("connecting to %s/events: %v", base, err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logErrorf("unexpected status: %s", resp.Status)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(resp.Body)

	var event, data string

	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case line == "":
			flush(event, data)
			event, data = "", ""

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
		logErrorf("reading event stream: %v", err)
		os.Exit(1)
	}

	logInfo("connection closed by server")
	os.Exit(0)
}

func send(base string) {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		msg := strings.TrimSpace(scanner.Text())
		if msg == "" {
			continue
		}

		if clientID == "" {
			logWarn("not connected yet, try again in a moment")
			continue
		}

		u := fmt.Sprintf("%s/chat/%s/%s", base, url.PathEscape(clientID), url.PathEscape(msg))

		resp, err := http.Get(u)
		if err != nil {
			logErrorf("send failed: %v", err)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			logErrorf("server returned %s", resp.Status)
		}
	}
}

func flush(event, data string) {
	if event == "" && data == "" {
		return
	}

	if event == "chat.id" && clientID == "" {
		clientID = data
		fmt.Printf("%sconnected as %s%s%s\n", colorGreen, colorBold, clientID, colorReset)
		fmt.Println("type a message and press enter to send it, ctrl+c to quit")
		return
	}

	fmt.Printf("%s[%s]%s %s\n", colorCyan, event, colorReset, data)
}

func logInfo(msg string) {
	log.Printf("%s%s%s", colorGray, msg, colorReset)
}

func logWarn(msg string) {
	log.Printf("%s%s%s", colorYellow, msg, colorReset)
}

func logErrorf(format string, args ...any) {
	log.Printf("%s"+format+"%s", append([]any{colorRed}, append(args, colorReset)...)...)
}
