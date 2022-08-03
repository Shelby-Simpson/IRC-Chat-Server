package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	Nickname      string
	Connection    *websocket.Conn
	GroupChats    []*GroupChat
	Send          chan []byte
	PersonalRooms []*PersonalRoom
}

// readPump listens for incoming messages
func (client *Client) readPump() {
	defer func() {
		for _, groupchat := range client.GroupChats {
			groupchat.Unregister <- client
		}
	}()
	var incomingJSON JSON
	for {
		_, incomingByteArray, err := client.Connection.ReadMessage()
		if err != nil {
			log.Printf("error occurred while reading message from client: %v", err)
			client.closeClient()
			break
		}
		// The incoming message is unmarshalled in a JSON struct
		err = json.Unmarshal(incomingByteArray, &incomingJSON)
		if err != nil {
			log.Printf("error occurred while unmarshalling JSON: %v", err)
			client.closeClient()
			break
		}

		switch incomingJSON.Type {
		// The client just connected to the server
		case "connect":
			// broadcast new user joined to all public group chats
			var request Request
			err = json.Unmarshal(incomingByteArray, &request)
			client.Nickname = request.SenderName
			if err != nil {
				log.Printf("error occurred while unmarshalling request: %v", err)
				client.closeClient()
				break
			}
			// Send list of existing groupchats
			var groupchatNamesCSV string
			for _, groupchat := range groupchats {
				groupchatNamesCSV += groupchat.Name + ","
			}
			groupchatNamesCSV = groupchatNamesCSV[:len(groupchatNamesCSV)-1] // removes last comma
			groupchatNamesResponse := Response{Type: "ChatRooms", Payload: groupchatNamesCSV}
			groupchatNamesJSON, err := json.Marshal(groupchatNamesResponse)
			if err != nil {
				log.Printf("error occurred while marshalling groupchat names: %v", err)
				client.closeClient()
				break
			}
			client.Send <- groupchatNamesJSON
			// Send list of existing clients
			var clientNamesCSV string
			for _, client := range clients {
				clientNamesCSV += client.Nickname + ","
			}
			clientNamesCSV = clientNamesCSV[:len(clientNamesCSV)-1] // removes last comma
			clientNamesResponse := Response{Type: "Clients", Payload: clientNamesCSV}
			clientNamesJSON, err := json.Marshal(clientNamesResponse)
			if err != nil {
				log.Printf("error occurred while marshalling client names: %v", err)
				client.closeClient()
				break
			}
			client.Send <- clientNamesJSON
			// Tell other clients about the new connection
			newClientResponse := Response{Type: "NewClient", Payload: client.Nickname}
			newClientJSON, err := json.Marshal(newClientResponse)
			if err != nil {
				log.Printf("error occurred while marshalling new client name: %v", err)
				client.closeClient()
				break
			}
			for _, c := range clients {
				if c != client {
					c.Send <- newClientJSON
				}
			}

		// The client is sending a message to a group chat
		case "groupchatmessage":
			var message Message
			err = json.Unmarshal(incomingByteArray, &message)
			if err != nil {
				log.Printf("error occurred while reading message from client: %v", err)
				client.closeClient()
				break
			}
			message.Timestamp = time.Now()
			messageJSON, err := json.Marshal(message)
			if err != nil {
				log.Printf("error occurred while marshalling message: %v", err)
				client.closeClient()
				break
			}
			for _, groupchat := range client.GroupChats {
				if groupchat.Name == message.ChatRoomName {
					groupchat.Broadcast <- messageJSON
				}
			}

		// The client is sending a message to a personal room
		case "personalroommessage":
			var message Message
			err = json.Unmarshal(incomingByteArray, &message)
			if err != nil {
				log.Printf("error occurred while reading message from client: %v", err)
				client.closeClient()
				break
			}
			message.Timestamp = time.Now()
			messageJSON, err := json.Marshal(message)
			if err != nil {
				log.Printf("error occurred while marshalling message: %v", err)
				client.closeClient()
				break
			}
			for _, personalroom := range client.PersonalRooms {
				if personalroom.Name == message.ChatRoomName {
					personalroom.Broadcast <- messageJSON
				}
			}

		// The client wants to create a new group chat
		case "creategroupchat":
			var groupchatRequest CreateGroupchatRequest
			err = json.Unmarshal(incomingByteArray, &groupchatRequest)
			if err != nil {
				log.Printf("error occurred while reading create groupchat request from client: %v", err)
				client.closeClient()
				break
			}
			newGroupChat := &GroupChat{
				Name:       groupchatRequest.Name,
				Broadcast:  make(chan []byte),
				Register:   make(chan *Client),
				Unregister: make(chan *Client),
				Clients:    make(map[*Client]bool),
			}
			go newGroupChat.run()
			groupchats = append(groupchats, newGroupChat)
			groupchatResponse := Response{
				Type:    "NewGroupChat",
				Payload: newGroupChat.Name,
			}
			groupchatResponseJSON, err := json.Marshal(groupchatResponse)
			if err != nil {
				log.Printf("error occurred while marshalling new groupchat response: %v", err)
				client.closeClient()
				break
			}
			for _, c := range clients {
				newGroupChat.Register <- c
				c.GroupChats = append(c.GroupChats, newGroupChat)
				c.Send <- groupchatResponseJSON
			}

		// The client wants to create a new personal room
		case "createpersonalroom":
			var personalroomRequest CreatePersonalRoomRequest
			err = json.Unmarshal(incomingByteArray, &personalroomRequest)
			if err != nil {
				log.Printf("error occurred while reading create personal room request from client: %v", err)
				client.closeClient()
				break
			}
			var client2 *Client
			for _, c := range clients {
				if c.Nickname == personalroomRequest.RecipientName {
					client2 = c
				}
			}
			newPersonalRoom := &PersonalRoom{
				Name:      client.Nickname + client2.Nickname,
				Client1:   client,
				Client2:   client2,
				Broadcast: make(chan []byte),
			}
			go newPersonalRoom.run()
			personalrooms = append(personalrooms, newPersonalRoom)
			personalroomResponse := Response{
				Type:    "NewPersonalRoom",
				Payload: newPersonalRoom.Name,
			}
			personalroomResponseJSON, err := json.Marshal(personalroomResponse)
			if err != nil {
				log.Printf("error occurred while marshalling new personal room response: %v", err)
				client.closeClient()
				break
			}
			client.PersonalRooms = append(client.PersonalRooms, newPersonalRoom)
			client.Send <- personalroomResponseJSON
			client2.PersonalRooms = append(client2.PersonalRooms, newPersonalRoom)
			client2.Send <- personalroomResponseJSON
		}
	}
}

// writePump writes messages to the client
func (client *Client) writePump() {
	defer func() {
		client.Connection.Close()
	}()
	for {
		message := <-client.Send
		err := client.Connection.WriteMessage(1, message)
		if err != nil {
			log.Printf("error occurred while writing message to client: %v", err)
			client.closeClient()
			break
		}
	}
}

// serveWS upgrades a connection to web socket and creates a new client
func serveWs(groupchats []*GroupChat, w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	connection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{GroupChats: groupchats, Connection: connection, Send: make(chan []byte)}
	clients = append(clients, client)
	for _, groupchat := range groupchats {
		groupchat.Register <- client
	}

	go client.readPump()
	go client.writePump()
}

// closeClient is abstracted code to remove a client when they disconnect
func (client *Client) closeClient() {
	client.Connection.Close()
	disconnectedClientResponse := Response{Type: "ClientDisconnect", Payload: client.Nickname}
	disconnectedClientJSON, _ := json.Marshal(disconnectedClientResponse)
	for i := range clients {
		if clients[i] == client {
			clients[i] = clients[len(clients)-1]
		} else {
			clients[i].Send <- disconnectedClientJSON
		}
	}
	clients = clients[:len(clients)-1]
}
