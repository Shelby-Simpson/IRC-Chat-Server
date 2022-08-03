package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// var chatrooms = make(map[string]ChatRoom)
var groupchats = make([]*GroupChat, 0)
var personalrooms = make([]*PersonalRoom, 0)
var clients = make([]*Client, 0)

// var personalRooms = make(map[string]PersonalRoom)
// var clients = make(map[string]bool)
// var broadcast = make(chan Message)

var upgrader = websocket.Upgrader{}

// type ChatRoom interface {
// 	createChatRoom(name string, client1 Client, client2 Client)
// 	broadcastMessage(message Message)
// }

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

type CreateGroupchatRequest struct {
	Name string
}

type CreatePersonalRoomRequest struct {
	RecipientName string
}

type Message struct {
	Message      string
	SenderName   string
	ChatRoomName string
	ChatRoomType string
	Timestamp    time.Time
}

// func (channel *GroupChat) createChatRoom(name string, client1 Client, client2 Client) {
// 	channel.Name = name
// 	channel.Clients = append(channel.Clients, &client1, &client2)
// }

// func (personalRoom *PersonalRoom) createChatRoom(name string, client1 Client, client2 Client) {
// 	personalRoom.Name = name
// 	personalRoom.Client1 = &client1
// 	personalRoom.Client2 = &client2
// }

// func (channel *GroupChat) broadcastMessage(message Message) {
// 	for {
// 		message := <-broadcast
// 		for client := range channel.Clients {
// 			err := client.Connection.WriteJSON(message)
// 			if err != nil {
// 				log.Printf("error occurred while writing message to client: %v", err)
// 				client.Connection.Close()
// 				delete(clients, client.Nickname)
// 				// remove client from channel Clients field
// 			}
// 		}
// 	}
// }

// func (personalRoom PersonalRoom) broadcastMessage(message Message) {
// 	for {
// 		message := <-broadcast
// 		err := personalRoom.Client1.Connection.WriteJSON(message)
// 		if err != nil {
// 			log.Printf("error occurred while writing message to client: %v", err)
// 			personalRoom.Client1.Connection.Close()
// 			delete(clients, personalRoom.Client1.Nickname)
// 			// delete personal room
// 		}
// 		err = personalRoom.Client2.Connection.WriteJSON(message)
// 		if err != nil {
// 			log.Printf("error occurred while writing message to client: %v", err)
// 			personalRoom.Client2.Connection.Close()
// 			delete(clients, personalRoom.Client2.Nickname)
// 			// delete personal room
// 		}
// 	}
// }

// func (channel *GroupChat) addClient(client Client) {
// 	channel.Clients = append(channel.Clients, &client)
// }

// {Part 1 }
// func HandleClients(w http.ResponseWriter, r *http.Request) {
// 	go broadcastMessage()
// 	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
// 	websocket, err := upgrader.Upgrade(w, r, nil)
// 	if err != nil {
// 		log.Fatal("error upgrading GET request to a websocket :: ", err)
// 	}
// 	defer websocket.Close()
// 	client := Client{
// 		Nickname:   "",
// 		Connection: websocket,
// 	}
// 	clients[client.Nickname] = true
// 	for channelName := range channels {
// 		channels[channelName].addClient(client)
// 	}
// 	var message Message
// 	for {
// 		err := websocket.ReadJSON(&message)
// 		if err != nil {
// 			log.Printf("error occurred while reading message : %v", err)
// 			delete(clients, client.Nickname)
// 			break
// 		}
// 		message.Timestamp = time.Now()
// 		broadcast <- message
// 	}
// }

// func broadcastMessage() {
// 	for {
// 		message := <-broadcast
// 		if channel, ok := channels[message.GroupChatName]; ok {
// 			channel.broadcastMessage(message)
// 		}
// 	}
// }

//Part 3
func main() {
	// Create 3 channels
	// channels = append(channels, &GroupChat{
	// 	Name:    "GroupChat1",
	// 	Clients: make(map[*Client]bool),
	// })
	// channels = append(channels, &GroupChat{
	// 	Name:    "GroupChat2",
	// 	Clients: make(map[*Client]bool),
	// })
	// channels = append(channels, &GroupChat{
	// 	Name:    "GroupChat3",
	// 	Clients: make(map[*Client]bool),
	// })
	// for _, channel := range channels {
	// 	go channel.run()
	// }

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

	// http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	http.ServeFile(w, r, "index.html")
	// })
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveWs(groupchats, w, r)
	})

	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("error starting http server :: ", err)

		return
	}
}

//Final Part Over
