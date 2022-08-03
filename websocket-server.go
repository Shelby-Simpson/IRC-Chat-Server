package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var groupchats = make([]*GroupChat, 0)
var personalrooms = make([]*PersonalRoom, 0)
var clients = make([]*Client, 0)

var upgrader = websocket.Upgrader{}

// Struct to unmarshal 'Type' property from incoming JSON message
type JSON struct {
	Type string
}

// Struct to unmarshal client requests other than
// sending a message in a chat room
type Request struct {
	Request    string
	SenderName string
}

// Struct to marshal server response into JSON
type Response struct {
	Type    string
	Payload string
}

// Struct to unmarshal client requests to
// create a new group chat
type CreateGroupchatRequest struct {
	Name string
}

// Struct to unmarshal client requests to
// create a new personal room
type CreatePersonalRoomRequest struct {
	RecipientName string
}

// Struct to marshal and unmarshal text messages
type Message struct {
	Message      string
	SenderName   string
	ChatRoomName string
	ChatRoomType string
	Timestamp    time.Time
}

func main() {
	// 3 default group chats
	groupchats = append(groupchats, &GroupChat{
		Name:       "GroupChat0",
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
	})
	groupchats = append(groupchats, &GroupChat{
		Name:       "GroupChat1",
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
	})
	groupchats = append(groupchats, &GroupChat{
		Name:       "GroupChat2",
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
	})

	for _, groupchat := range groupchats {
		go groupchat.run()
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveWs(groupchats, w, r)
	})

	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("error starting http server :: ", err)

		return
	}
}
