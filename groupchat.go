package main

// Implements ChatRoom
type GroupChat struct {
	Name       string
	Clients    map[*Client]bool
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
}

func (groupchat *GroupChat) run() {
	for {
		select {
		case client := <-groupchat.Register:
			groupchat.Clients[client] = true
		case client := <-groupchat.Unregister:
			delete(groupchat.Clients, client)
			// close(client.Send)
		case message := <-groupchat.Broadcast:
			for client := range groupchat.Clients {
				select {
				case client.Send <- message:
				default:
					// close(client.Send)
					delete(groupchat.Clients, client)
				}
			}
		}
	}
}
