# Instructions to Run the Server
1. Run 'go run websocket' in the server's directory

# Instructions to Begin Using the Client
1. Run 'npm start' in the client's directory.
2. Connect to localhost:3006 in the browser.
3. Enter a nickname and click 'submit'.
4. Open another instance of the client by connecting to localhost:3006 in another browser tab.
5. Enter a nickname for the second client and click 'submit'.

# Instructions to Type in a Chat Room
1. On one client, click on an existing chatroom and type a message.
2. On the other client, open that chatroom to see the message and type one back!

# Instructions to Create a New Group Chat
1. On one client, enter a name in the textbox and click 'New Group Chat'.
2. The new group chat should appear in the list of chat rooms.

# Instructions to Create a New Personal Room
1. On one client, select a client from the dropdown and click 'New Direct Message'.
2. The new personal room should appear in the list of chat rooms.



# Existing Issues with the Application:
1. Client data is not persistent.  When a client exits the browser tab or refreshes, they disconnect from the server and all of their data is lost.  A database (such as MongoDB) will be added in the future to fix this issue.
2. All chat rooms are stored on one server.  Clients cannot select a different port to connect to a different one.
3. A client in unable to delete a chat room.  Personal rooms need to be deleted when a client disconnects.
4. Nicknames and group chat names can be empty strings and they are not unique.  Adding name checks on both the frontend and backend can fix this issue.
5. The frontend is not responsive.  Certain screens will not be formatted well.
6. The general style of the frontend is not very user-friendly.
