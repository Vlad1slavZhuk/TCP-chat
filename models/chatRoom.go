package models

import (
	"fmt"
)

type ChatRoom struct {
	name     string
	users    []*User
	messages []string
}

func NewChatRoom(name string) *ChatRoom {
	return &ChatRoom{
		name:     name,
		users:    make([]*User, 0),
		messages: make([]string, 0),
	}
}

func (chatRoom *ChatRoom) SendAll(msg string) {
	chatRoom.messages = append(chatRoom.messages, msg)

	for _, user := range chatRoom.users {
		user.outgoing <- msg
	}
}

func (chatRoom *ChatRoom) Join(u *User) {
	u.chatRoom = chatRoom

	for _, message := range chatRoom.messages {
		u.outgoing <- message
	}

	chatRoom.users = append(chatRoom.users, u)
	chatRoom.SendAll(fmt.Sprintf("<%s> joined the chat room.\n", u.name))
}

func (chatRoom *ChatRoom) Leave(u *User) {
	chatRoom.SendAll(fmt.Sprintf("<%s> left this chat room.\n", u.name))

	for i, otherUser := range chatRoom.users {
		if u == otherUser {
			chatRoom.users = append(chatRoom.users[:i], chatRoom.users[i:]...)
			break
		}
	}
	u.chatRoom = nil
}

func (chatRoom ChatRoom) Delete() {
	if len(chatRoom.users) > 0 {
		chatRoom.SendAll("Chat room is inactive and begin delete.\n")
		for _, user := range chatRoom.users {
			user.chatRoom = nil
		}
	}
}
