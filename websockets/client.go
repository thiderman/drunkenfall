package websockets

import (
	"fmt"
	"io"
	"log"

	"golang.org/x/net/websocket"
)

const channelBufSize = 100

var maxID int

// Client is the representation of a listener.
type Client struct {
	id     int
	ws     *websocket.Conn
	server *Server
	ch     chan *Message
	doneCh chan bool
	pingCh chan bool
}

// NewClient creates a new chat client.
func NewClient(ws *websocket.Conn, server *Server) *Client {
	if ws == nil {
		panic("ws cannot be nil")
	}

	if server == nil {
		panic("server cannot be nil")
	}

	maxID++
	ch := make(chan *Message, channelBufSize)
	doneCh := make(chan bool)
	pingCh := make(chan bool)

	return &Client{maxID, ws, server, ch, doneCh, pingCh}
}

// Conn returns the websocket connection
func (c *Client) Conn() *websocket.Conn {
	return c.ws
}

// Write sends a message or drops the connection
func (c *Client) Write(msg *Message) {
	select {
	case c.ch <- msg:
	default:
		c.server.Del(c)
		err := fmt.Errorf("client %d is disconnected", c.id)
		c.server.Err(err)
	}
}

// Done closes the client
func (c *Client) Done() {
	c.doneCh <- true
}

// Ping sends a ping message to the client
func (c *Client) Ping() {
	c.pingCh <- true
}

// Listen writes and reads request via channel
func (c *Client) Listen() {
	go c.listenWrite()
	c.listenRead()
}

// messageError handles when a message could not properly be sent.
func (c *Client) messageError() {
	log.Print("Error in send. Disconnecting.")
	c.Done()
}

func (c *Client) listenWrite() {
	// log.Println("Listening write to client")
	for {
		select {

		// send message to the client
		case msg := <-c.ch:
			if err := websocket.JSON.Send(c.ws, msg); err != nil {
				c.messageError()
				return
			}

		// send ping to the client
		case <-c.pingCh:
			if err := websocket.JSON.Send(c.ws, Ping{P: 1}); err != nil {
				c.messageError()
				return
			}

		// receive done request
		case <-c.doneCh:
			c.server.Del(c)
			c.doneCh <- true // for listenRead method
			return
		}
	}
}

// Listen read request via chanel
func (c *Client) listenRead() {
	// log.Println("Listening read from client")
	for {
		select {

		// receive done request
		case <-c.doneCh:
			c.server.Del(c)
			c.doneCh <- true // for listenWrite method
			return

		// read data from websocket connection
		default:
			var msg Message
			err := websocket.JSON.Receive(c.ws, &msg)
			if err == io.EOF {
				c.doneCh <- true
			} else if err != nil {
				c.server.Err(err)
			} else {
				c.server.SendAll(&msg)
			}
		}
	}
}
