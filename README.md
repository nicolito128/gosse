# Gosse

Package to handle server-sent events with Go.

## Getting started

Install the package using `go get`:

```bash
go get github.com/nicolito128/gosse
```

Import the package in your Go code:

```go
import "github.com/nicolito128/gosse"
```

## Quick example

```go
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
	c.Send([]byte(msg.String()))

	<-c.Done()
}
```

Connect with `curl -N http://localhost:8080/events` and you'll see:

```
event: greeting
retry: 3000
data: hello, world!
```

## Links

* [Using server-sent events - Web APIs | MDN](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events)
