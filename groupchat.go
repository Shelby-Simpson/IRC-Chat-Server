package main

// Implements ChatRoom
type GroupChat struct {
	Name       string
	Clients    map[*Client]bool
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
}

// Implements ChatRoom
type PersonalRoom struct {
	Name      string
	Client1   *Client
	Client2   *Client
	Broadcast chan []byte
}

// run listens for clients registering, unregistering,
// and broadcasting messages on a group chat
func (groupchat *GroupChat) run() {
	for {
		select {
		case client := <-groupchat.Register:
			groupchat.Clients[client] = true
		case client := <-groupchat.Unregister:
			delete(groupchat.Clients, client)
		case message := <-groupchat.Broadcast:
			for client := range groupchat.Clients {
				select {
				case client.Send <- message:
				default:
					delete(groupchat.Clients, client)
				}
			}
		}
	}
}

// run broadcasts messages to the clients in a personal room
func (personalroom *PersonalRoom) run() {
	for {
		message := <-personalroom.Broadcast
		personalroom.Client1.Send <- message
		personalroom.Client2.Send <- message
	}
}
