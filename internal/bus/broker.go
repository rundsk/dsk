// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bus

import "log"

func NewBroker() *Broker {
	return &Broker{
		Subscribable: &Subscribable{},
		incoming:     make(chan *Message, 10),
		done:         make(chan bool),
		topics:       make([]string, 0),
	}
}

// Broker is the main event bus that services inside the DSK
// backend subscribe to.
type Broker struct {
	*Subscribable

	// Incoming messages are sent here.
	incoming chan *Message

	// Quit channel, receiving true, when de-initialized.
	done chan bool

	topics []string
}

func (b *Broker) Start() error {
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
	return nil
}

func (b *Broker) Stop() error {
	b.done <- true
	return nil
}

func (b *Broker) Close() error {
	b.UnsubscribeAll()
	return nil
}

// Accept a message for fan-out. Will never block. When the
// buffer is full the message will be discarded and not delivered.
func (b *Broker) Accept(m *Message) (ok bool) {
	select {
	case b.incoming <- m:
		ok = true
	default:
		log.Printf("Message buffer full, discarded: %s", m)
		ok = false
	}
	return
}

func (b *Broker) RegisterTopic(name string) {
	b.topics = append(b.topics, name)
}
