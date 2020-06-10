package main

import (
	"log"

	"github.com/gorilla/websocket"
	r "gopkg.in/rethinkdb/rethinkdb-go.v6"
)

type FindHandler func(string) (Handler, bool)

type Client struct {
	send         chan Message
	socket       *websocket.Conn
	findHandler  FindHandler
	session      *r.Session
	stopChannels map[int]chan bool
	id           string
	userName     string
}

// Method to create a new stop for the goroutine
func (c *Client) NewStopChannel(stopKey int) chan bool {
	// remove existing identical channel
	c.StopForKey(stopKey)
	stop := make(chan bool)
	c.stopChannels[stopKey] = stop
	return stop
}

// Method to delete an existing stop channel
func (c *Client) StopForKey(key int) {
	if ch, found := c.stopChannels[key]; found {
		// make sure that the channel endpoint is stopped before plucking off
		ch <- true
		// delete the channel from the hashmap
		delete(c.stopChannels, key)
	}
}

// Method to read the input from the client
func (client *Client) Read() {
	var message Message
	for {
		if err := client.socket.ReadJSON(&message); err != nil {
			break
		}
		if handler, found := client.findHandler(message.Name); found {
			handler(client, message.Data)
		}
	}
	client.socket.Close()
}

// Method to write the output to the client
func (client *Client) Write() {
	for msg := range client.send {
		if err := client.socket.WriteJSON(msg); err != nil {
			break
		}
	}
	client.socket.Close()
}

// Method to close the client connection
func (c *Client) Close() {
	// stop all the goroutine running
	for _, ch := range c.stopChannels {
		ch <- true
	}
	// close the socket
	close(c.send)
	// delete user
	r.Table("user").Get(c.id).Delete().Exec(c.session)
}

// Create a new client
func NewClient(socket *websocket.Conn, findHandler FindHandler,
	session *r.Session) *Client {
	var user User
	user.Name = "anonymous"
	res, err := r.Table("user").Insert(user).RunWrite(session)
	if err != nil {
		log.Println(err.Error())
	}
	var id string
	if len(res.GeneratedKeys) > 0 {
		id = res.GeneratedKeys[0]
	}
	return &Client{
		send:         make(chan Message),
		socket:       socket,
		findHandler:  findHandler,
		session:      session,
		stopChannels: make(map[int]chan bool),
		id:           id,
		userName:     user.Name,
	}
}
