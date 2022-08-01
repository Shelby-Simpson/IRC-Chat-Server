package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	Nickname   string
	Connection *websocket.Conn
	GroupChats []*GroupChat
	Send       chan []byte
	// PersonalRooms []*PersonalRoom
}

func (client *Client) readPump() {
	defer func() {
		for _, groupchat := range client.GroupChats {
			groupchat.Unregister <- client
		}
	}()
	var incomingJSON JSON
	for {
		_, incomingByteArray, err := client.Connection.ReadMessage()
		log.Print("Read message type")
		if err != nil {
			log.Printf("error occurred while reading message from client: %v", err)
			client.Connection.Close()
			break
		}
		err = json.Unmarshal(incomingByteArray, &incomingJSON)
		if err != nil {
			log.Printf("error occurred while unmarshalling JSON: %v", err)
			client.Connection.Close()
			break
		}

		switch incomingJSON.Type {

		case "connect":
			// broadcast new user joined to all public group chats
			var request Request
			err = json.Unmarshal(incomingByteArray, &request)
			// log.Print(incomingJSON)
			// log.Print(request)
			// client.Nickname = request.SenderName
			if err != nil {
				log.Printf("error occurred while unmarshalling request: %v", err)
				client.Connection.Close()
				break
			}
			var groupchatNamesCSV string
			for _, groupchat := range groupchats {
				groupchatNamesCSV += groupchat.Name + ","
			}
			// remove last comma
			groupchatNamesCSV = groupchatNamesCSV[:len(groupchatNamesCSV)-1]
			response := Response{Type: "GroupChats", Payload: groupchatNamesCSV}
			groupchatNamesJSON, err := json.Marshal(response)
			if err != nil {
				log.Printf("error occurred while marshalling groupchat names: %v", err)
				client.Connection.Close()
				break
			}
			client.Send <- groupchatNamesJSON

		case "message":
			var message Message
			err = json.Unmarshal(incomingByteArray, &message)
			if err != nil {
				log.Printf("error occurred while reading message from client: %v", err)
				client.Connection.Close()
				break
			}
			message.Timestamp = time.Now()
			messageJSON, err := json.Marshal(message)
			if err != nil {
				log.Printf("error occurred while marshalling message: %v", err)
				client.Connection.Close()
				break
			}
			for _, groupchat := range client.GroupChats {
				if groupchat.Name == message.GroupChatName {
					groupchat.Broadcast <- messageJSON
				}
			}

		case "creategroupchat":
			var groupchatRequest CreateGroupchatRequest
			err = json.Unmarshal(incomingByteArray, &groupchatRequest)
			if err != nil {
				log.Printf("error occurred while reading create groupchat request from client: %v", err)
				client.Connection.Close()
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
				client.Connection.Close()
				break
			}
			for _, c := range clients {
				newGroupChat.Register <- c
				c.GroupChats = append(c.GroupChats, newGroupChat)
				c.Send <- groupchatResponseJSON
			}
		}
	}
}

func (client *Client) writePump() {
	defer func() {
		client.Connection.Close()
	}()
	for {
		message := <-client.Send
		err := client.Connection.WriteMessage(1, message)
		if err != nil {
			log.Printf("error occurred while writing message to client: %v", err)
			client.Connection.Close()
			break
		}
	}
}

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
