package models

import (
	"fmt"
	"log"
	"strings"
)

type Lobby struct {
	users          []*User
	chatRooms      map[string]*ChatRoom
	incoming       chan *Message
	join           chan *User
	leave          chan *User
	deleteChatRoom chan *ChatRoom
}

func NewLobby() *Lobby {
	lobby := &Lobby{
		users:          make([]*User, 0),
		chatRooms:      make(map[string]*ChatRoom),
		incoming:       make(chan *Message, 5),
		join:           make(chan *User, 5),
		leave:          make(chan *User, 5),
		deleteChatRoom: make(chan *ChatRoom, 5),
	}
	lobby.Listen()
	return lobby
}

func (l *Lobby) Listen() {
	go func() {
		for {
			select {
			case msg := <-l.incoming:
				l.Parse(msg)
			case user := <-l.join:
				l.Join(user)
			case user := <-l.leave:
				l.Leave(user)
			}
		}
	}()
}

// Parsing strings and executing commands
func (l *Lobby) Parse(msg *Message) {
	args := strings.TrimSpace(msg.text)
	arrStr := strings.SplitN(args, " ", 2)

	command := arrStr[0]
	text := ""
	if len(arrStr) == 2 {
		text = arrStr[1]
	}

	switch command {
	case "/help":
		l.Help(msg.user)
	case "/list":
		l.ListChatRooms(msg.user)
	case "/create":
		l.CreateChatRoom(msg.user, text)
	case "/join":
		l.JoinToChatRoom(msg.user, text)
	case "/leave":
		l.LeaveChatRoom(msg.user)
	case "/name":
		l.ChangeName(msg.user, text)
	case "/quit":
		msg.user.Quit()
	case "/del":
		l.DeleteChatRoom(msg.user, text)
	default:
		l.SendMessage(msg)
	}
}

func (l *Lobby) Join(user *User) {
	// If the number of users is overflowed, disabled log out of the client
	if len(l.users) > 50 {
		user.Quit()
		return
	}

	l.users = append(l.users, user)
	log.Printf("New user (%v) joined in lobby.\n", user.conn.RemoteAddr().String())
	user.outgoing <- "Welcome to the Chat! Type \"/help\" to get a list commands.\n"

	go func() {
		for message := range user.incoming {
			l.incoming <- message
		}
		l.leave <- user
	}()
}

func (l *Lobby) Leave(user *User) {
	if user.chatRoom != nil {
		user.chatRoom.Leave(user)
	}

	for i, otherUser := range l.users {
		if user == otherUser {
			l.users = append(l.users[:i], l.users[i:]...)
			break
		}
	}
	close(user.outgoing)
	log.Printf("Closed %s's outgoing channel.\n", user.name)
}

func (l *Lobby) DeleteChatRoom(user *User, name string) {
	if l.chatRooms[name] != nil && l.chatRooms[name].author == user.name {
		if user.chatRoom == nil {
			l.chatRooms[name].Delete()
			delete(l.chatRooms, l.chatRooms[name].name)
		} else {
			if user.chatRoom.name != name {
				l.LeaveChatRoom(user)
			}
			l.chatRooms[name].Delete()
			delete(l.chatRooms, l.chatRooms[name].name)
		}
		user.outgoing <- fmt.Sprintf("<SUCCESS> You delete chat room \"%s\".\n", name)
		log.Printf("Deleted chat room by \"%s\".", user.name)
	} else {
		user.outgoing <- fmt.Sprintf("<WARNING> You are not the creator of the chat room \"%s\".\n", name)
		log.Printf("An attempt to delete the chat room by user \"%s\".\n", user.name)
	}
}

func (l *Lobby) SendMessage(msg *Message) {
	if msg.user.chatRoom == nil {
		msg.user.outgoing <- "<Error> You cannot send message here. Create chat room and enter.\n"
		log.Printf("%s tried to send message in room.\n", msg.user.name)
		return
	}

	msg.user.chatRoom.SendAll(msg.String())
	log.Printf("%s sent message\n", msg.user.name)
}

func (l *Lobby) Help(user *User) {
	user.outgoing <- "\n\tCommands:\n"
	user.outgoing <- "\t/help - lists all commands.\n"
	user.outgoing <- "\t/list - lists all chat room.\n"
	user.outgoing <- "\t/create foo - creates a chat room with name \"foo\".\n"
	user.outgoing <- "\t/del foo - deletes a chat room.\n"
	user.outgoing <- "\t/join foo - joins a chat room named foo.\n"
	user.outgoing <- "\t/leave - leaves the current chat room.\n"
	user.outgoing <- "\t/name foo - changes your name to foo.\n"
	user.outgoing <- "\t/quit - quits the program.\n\n"
	log.Printf("%s requested help.", user.name)
}

func (l *Lobby) ListChatRooms(user *User) {
	if len(l.chatRooms) == 0 {
		user.outgoing <- "\n\tChat rooms: nil\n"
	} else {
		user.outgoing <- "\n\tChat Rooms:\n"
		for name := range l.chatRooms {
			user.outgoing <- "\t" + name + "\n"
		}
		user.outgoing <- "\n"
	}

	log.Printf("%s listed chat rooms.", user.name)
}

func (l *Lobby) CreateChatRoom(user *User, name string) {
	if l.chatRooms[name] != nil {
		user.outgoing <- "<WARNING> A chat room with that name already exists.\n"
		log.Printf("%s tried to create chat room with a name already in use.\n", user.name)
		return
	}
	if user.name == "Anonymous" {
		user.outgoing <- "<WARNING> Change name and try again.\n"
		return
	}

	chatRoom := NewChatRoom(name, user.name)
	l.chatRooms[name] = chatRoom

	user.outgoing <- fmt.Sprintf("<SUCCESS> Created chat room \"%s\".\n", chatRoom.name)
	log.Printf("%s created chat room \"%s\".", user.name, chatRoom.name)
}

func (l *Lobby) JoinToChatRoom(user *User, name string) {
	if l.chatRooms[name] == nil {
		user.outgoing <- "<WARNING> A chat room with that mame does not exists.\n"
		log.Printf("%s tried to join a chat room that does not exists.\n", user.name)
		return
	}

	if user.chatRoom != nil {
		if user.chatRoom.name == name {
			user.outgoing <- fmt.Sprintf("<WARNING> You are already in the chat room with the name \"%s\".\n", name)
			return
		}
		l.LeaveChatRoom(user)
	}

	l.chatRooms[name].Join(user)
	log.Printf("%s joined to chat room \"%s\".\n", user.name, name)
}

func (l *Lobby) LeaveChatRoom(user *User) {
	if user.chatRoom == nil {
		user.outgoing <- "<WARNING> You cannot leave lobby.\n"
		log.Printf("%s tried to leave the room.\n", user.name)
		return
	}
	user.chatRoom.Leave(user)
}

func (l *Lobby) ChangeName(user *User, name string) {
	if user.chatRoom == nil {
		user.outgoing <- fmt.Sprintf("<SUCCESS> Change name to %s.\n", name)
	} else {
		user.chatRoom.SendAll(
			fmt.Sprintf("<NOTICE> %s changed their name to %s", user.name, name))
	}
	log.Printf("%s changed their name to %s", user.name, name)
	user.name = name
}
