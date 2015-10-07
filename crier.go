package crier

import (
	"net/http"
	"fmt"
	"log"
	"errors"
)

type Crier struct {
	// Listing of everyone connected to this crier
	audience   map[*listener]chan bool
	// Announce to the entire audience
	broadcast  chan string
	// Groups which listeners can join
	groups     map[string][]*listener
	// Send a message to all listeners within a channel
	publish    chan map[string]string
	// Send a message to a specific listener
	whisper    chan map[*listener]string
	// Add a member to the audience
	introduce  chan *listener
	// Remove a member from the audience
	leave      chan *listener
	// Groups are automatically created/deleted based upon need
	// Add a listener to a group
	add        chan map[string]*listener
	// Remove a listener from a group
	remove     chan map[string]*listener
	// Forcefully disconnect a client
	disconnect chan *listener


	// Stop the crier
	shutdown   chan bool

	history    *history
}

func NewCrier() *Crier {
	c := &Crier{
		audience:        make(map[*listener]chan bool),
		broadcast:        make(chan string),
		groups:            make(map[string][]*listener),
		publish:        make(chan map[string]string),
		whisper:        make(chan map[*listener]string),
		introduce:        make(chan *listener),
		leave:            make(chan *listener),
		add:            make(chan map[string]*listener),
		remove:            make(chan map[string]*listener),
		shutdown:        make(chan bool),
		disconnect:        make(chan *listener),
		history:        &history{
			size: 20,
			history: make(map[int]*event),
			lastId: 0,
		},
	}
	go c.start()
	return c
}

func (c *Crier) start() {
	// The loop used to perform actions when information is sent over a channel
	for {
		select {
		// Add a new client
		case client := <-c.introduce:
			c.audience[client] = make(chan bool)
		// Remove a client
		case l := <-c.leave:
			delete(c.audience, l)
		// Broadcast to all clients
		case m := <-c.broadcast:
			c.history.addBroadcast(m)
			for client, _ := range c.audience {
				client.send(m)
			}
		//	Send a message to a group
		case publish := <-c.publish:
			for m, g := range publish {
				c.history.addPublish(m, g)
				for _, client := range c.groups[g] {
					client.send(m)
				}
			}
		// Send message to a specific client
		case clients := <-c.whisper:
			for client, m := range clients {
				client.send(m)
			}
		// Add a client to a group
		case groups := <-c.add:
			for g, client := range groups {
				if ok := c.groups[g]; ok != nil {
					c.groups[g] = make([]*listener, 2)
				}
				c.groups[g] = append(c.groups[g], client)
			}
		// Remove a client from a group
		case groups := <-c.remove:
			for g, client := range groups {
				if ok := c.groups[g]; ok != nil {
					// the group doesn't exist
					// TODO(morgan): silently failing is bad.
					continue
				}
				i, err := c.findListenerInGroup(client, g);
				if err != nil {
					// client is not in that group
					// TODO(morgan): silently failing is bad.
					continue
				}
				gp := c.groups[g]
				gp[i] = gp[len(gp) - 1]
				gp = gp[0:len(gp) - 1]
			}
		case d := <- c.disconnect:
			c.audience[d] <- true
//		case s := <- c.shutdown:
//			c.audience[s].disconnect <- true
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

func (c *Crier) Disconnect(w http.ResponseWriter) {
	for l, ch := range c.audience {
		if l.conn == w {
			// Send true to the disconnect channel
			ch <- true
		}
	}
}

func (c *Crier) Shutdown() {
	//Shutdown the crier
}

func (c *Crier) findListenerInGroup(l *listener, group string) (int, error) {
	for i := 0; i < len(c.groups[group]); i++ {
		if (c.groups[group][i] == l) {
			return i, nil
		}
	}
	return 0, errors.New("Listener not found.")
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
		left : make(chan bool),
		disconnect: make(chan bool),
	}

	go func() {
		<-notify
		log.Println("Disconnected.")
		// Remove this client from the map of attached clients when `EventHandler` exits.
		c.leave <- l
		l.left <- true
	}()
	c.audience[l] = l.left
	log.Println("Connected.")

	//sit here and wait
	<-l.left
	log.Print("End of ServeHTTP.")
}

type listener struct {
	conn       http.ResponseWriter
	// Forceful disconnect
	disconnect chan bool
	// Client left
	left       chan bool
}

func (l *listener) send(message string) {
	f, _ := l.conn.(http.Flusher)
	fmt.Fprint(l.conn, message)
	f.Flush()
	log.Print("Message sent via sender.")
}
