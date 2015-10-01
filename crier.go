package crier

import (
	"net/http"
	"fmt"
	"log"
)

type Crier struct {
	// Listing of everyone connected to this crier
	audience  map[*listener]bool
	// Announce to the entire audience
	broadcast chan string
	// Groups which listeners can join
	groups    map[string]*listener
	// Send a message to all listeners within a channel
	publish   chan map[string]string
	// Send a message to a specific listener
	whisper   chan map[*listener]string
	// Add a member to the audience
	introduce chan *listener
	// Remove a member from the audience
	leave     chan *listener
	// Add a listener to a group
	add       chan map[string]*listener
	// Remove a listener from a group
	remove    chan map[string]*listener
	// Create a group
	create    chan string
	// Destroy a group
	destroy   chan string
	// Stop the crier
	shutdown  chan bool

	history	*history
}

func NewCrier() *Crier {
	c := &Crier{
		audience:	make(map[*listener]bool),
		broadcast:	make(chan string),
		groups:		make(map[string]*listener),
		publish:	make(chan map[string]string),
		whisper:	make(chan map[*listener]string),
		introduce:	make(chan *listener),
		leave:		make(chan *listener),
		add:		make(chan map[string]*listener),
		remove:		make(chan map[string]*listener),
		create:		make(chan string),
		destroy:	make(chan string),
		shutdown:	make(chan bool),
		history:	&history {
			size: 20,
			history: make(map[int]*event),
			lastId: 0,
		},
	}
	go c.start()
	return c
}

func (c *Crier) start() {
	for {
		select {
		// Add a new client
		case client := <- c.introduce:
			c.audience[client] = true
		case m := <-c.broadcast:
			c.history.addBroadcast(m)
			for client, _ := range c.audience {
				client.send(m)
			}
		}
	}
}

func (c *Crier) Broadcast(message string) {
	c.broadcast <- message
}

func (c *Crier) Publish(message, group string) {
	c.publish <- map[string]string{
		group: message,
	}
}

// Need to abstract out *listener possibly?
func (c *Crier) Whisper(message string, target *listener) {
	c.whisper <- map[*listener]string{
		target: message,
	}
}

func (c *Crier) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// Make sure that the writer supports flushing.
	_, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming is not supported by your browser.", http.StatusInternalServerError)
		return
	}
	// Set the headers related to event streaming.
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Listen to the closing of the http connection via the CloseNotifier
	notify := w.(http.CloseNotifier).CloseNotify()

	l := &listener{
		conn: w,
	}

	left := make(chan bool)

	go func() {
		<-notify
		log.Println("Disconnected.")
		// Remove this client from the map of attached clients when `EventHandler` exits.
		c.leave <- l
		left <- true
	}()

	c.audience[l] = true
	log.Println("Connected.")

	for {
		//sit here and wait
		if true == <-left {
			break
		}
	}
	log.Println("End of ServeHTTP")
}

type listener struct {
	conn http.ResponseWriter
}

func (l *listener) send(message string) {
	f, _ := l.conn.(http.Flusher)
	fmt.Fprint(l.conn, message)
	f.Flush()
	log.Print("Message sent via sender.")
}
