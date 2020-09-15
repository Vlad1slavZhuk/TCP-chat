package models

import (
	"fmt"
	"time"
)

type Message struct {
	time time.Time
	user *User
	text string
}

func NewMessage(time time.Time, user *User, text string) *Message {
	return &Message{
		time: time,
		user: user,
		text: text,
	}
}

func (msg *Message) String() string {
	return fmt.Sprintf("%v [%v]: %v", msg.user.name, msg.time.Format("15:04:00"), msg.text)
}
