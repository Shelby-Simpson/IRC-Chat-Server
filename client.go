package main

import (
	"json"
	"log"
	"net/http"
	// "regexp"
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
		client.Connection.Close()
	}()
	var incomingJSON JSON
	for {
		err := client.Connection.ReadJSON(&incomingJSON)
		if err != nil {
			log.Printf("error occurred while writing message to client: %v", err)
			client.Connection.Close()
			break
		}
		// Check the type and act accordingly
		// messageType := getType(string(messageByteArray))
		switch incomingJSON.Type {
		case "connect":
			// broadcast new user joined to all public group chats
			// send list of group chats to this client
			var request Request
			err = client.Connection.ReadJSON(&request)

		case "message":
			var message Message
			err = client.Connection.ReadJSON(&message)
			message.Timestamp = time.Now()
			messageJSON := json.Marshall(message)
			for _, groupchat := range client.GroupChats {
				if groupchat.Name == message.GroupChatName {
					groupchat.Broadcast <- &messageJSON
				}
			}
		}

	}
}

// // Get message type from web socket message
// func getType(s string) string {
// 	re := regexp.MustCompile(`"type":"[a-zA-Z0-9 ]*"`)
// 	messageType := re.FindString(s)
// 	// cut off "type":"<message type>" to get <message type>
// 	return string([]rune(messageType)[8 : len([]rune(messageType))-1])
// }

func (client *Client) writePump() {
	defer func() {
		client.Connection.Close()
	}()
	for {
		message := <-client.Send
		err := client.Connection.WriteJSON(message)
		if err != nil {
			log.Printf("error occurred while writing message to client: %v", err)
			client.Connection.Close()
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
	for _, groupchat := range groupchats {
		groupchat.Register <- client
	}

	go client.readPump()
	go client.writePump()
}
