// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bus

import (
	"fmt"
	"math/rand"
)

func NewMessage(topic string, text string) *Message {
	return &Message{
		Id:    rand.Int(),
		Topic: topic,
		Text:  text,
	}
}

type Message struct {
	Id    int
	Topic string
	Text  string
}

func (m *Message) String() string {
	return fmt.Sprintf("<Message %d %s>%s</Message>", m.Id, m.Topic, m.Text)
}
