/*
Package gosse helps you send Server-Sent Events (SSE) from a Go HTTP server.

Server-Sent Events let a server push messages to a browser over a single,
long-lived HTTP connection. It's a simpler alternative to WebSockets when you
only need one-way communication, from server to client.

# Basic usage

Call Upgrade inside an http.HandlerFunc to turn the connection into an SSE
stream. This returns a *Channel, which you use to send messages to that
one client:

	func handleEvents(w http.ResponseWriter, r *http.Request) {
		c := gosse.Upgrade(w, r)
		defer c.Close()

		msg := gosse.NewMessage([]byte("hello, goose!"), "greeting", 0)
		c.Send(msg.Bytes())

		<-c.Done()
	}

Each client that connects gets its own Channel, so you'll typically call
Upgrade once per incoming request.

# Messages

A Message represents one event on the wire. Use NewMessage to build one,
then send its bytes through a Channel:

	msg := gosse.NewMessage(data, "event-name", 0)
	c.Send(msg.Bytes())

The last argument to NewMessage is the retry delay, in milliseconds, that
tells the browser how long to wait before reconnecting if the connection
drops. Passing any value less than 0 uses DefaultRetry.

# Sending vs. writing

Send queues a message to be delivered asynchronously; it's safe to call from
any goroutine and won't block on the network. Write, used internally by
Upgrade, writes directly to the client and flushes the response. Most code
should just use Send.

# Closing and errors

A Channel stays open until you call Close, the client disconnects, or an
error occurs. Done returns a channel you can wait on to know when that
happens:

	<-c.Done()

After the Channel is closed, check Error to see why, if you need to know.

# Listening for messages

If you'd rather react to each message than send them from your own
goroutine, use Listen. It blocks and calls the given function for every
message received, until the Channel closes:

	err := c.Listen(func(msg []byte) {
		log.Println("sent:", string(msg))
	})

# Configuration

Upgrade and NewChannel accept optional ChannelOpt values to configure the
Channel. For example, WithBufferSize changes how many messages can be
queued before Send blocks:

	c := gosse.Upgrade(w, r, gosse.WithBufferSize(4096))
*/
package gosse
