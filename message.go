// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"math/rand"
)

const (
	// MessageTypeTreeSynced happens whenever the node tree has
	// been (initially or after a rsync) synchronized.
	MessageTypeTreeSynced = "tree-synced"

	// MessageTypeTreeChanged happens when something in the tree has
	// been changed and a resync needs to happen.
	MessageTypeTreeChanged = "tree-changed"
)

func NewMessage(typ string, text string) *Message {
	return &Message{
		id:   rand.Int(),
		typ:  typ,
		text: text,
	}
}

type Message struct {
	id   int
	typ  string
	text string
}

func (m *Message) String() string {
	return fmt.Sprintf("<Message %d %s>%s</Message>", m.id, m.typ, m.text)
}

func NewMessageBroker() *MessageBroker {
	return &MessageBroker{
		Subscribable: &Subscribable{},
		incoming:     make(chan *Message, 10),
		done:         make(chan bool),
	}
}

type MessageBroker struct {
	*Subscribable

	// Incoming messages are sent here.
	incoming chan *Message

	// Quit channel, receiving true, when de-initialized.
	done chan bool
}

func (b *MessageBroker) Start() {
	go func() {
		for {
			select {
			case m := <-b.incoming:
				b.NotifyAll(m)
			case <-b.done:
				log.Print("Message broker is closing...")
				return
			}
		}
	}()
}

func (b *MessageBroker) Close() {
	b.UnsubscribeAll()
	b.done <- true
}

// Accept a message for fan-out. Will never block. When the
// buffer is full the message will be discarded and not delivered.
func (b *MessageBroker) Accept(m *Message) (ok bool) {
	select {
	case b.incoming <- m:
		ok = true
	default:
		log.Printf("Message buffer full, discarded: %s", m)
		ok = false
	}
	return
}
