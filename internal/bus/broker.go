// Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bus

import (
	"fmt"
	"log"
)

func NewBroker() (*Broker, error) {
	log.Print("Initializing message broker...")

	b := &Broker{
		Subscribable: &Subscribable{},
		incoming:     make(chan *Message, 10),
		done:         make(chan bool),
	}
	return b, b.Open()
}

// Broker is the main event bus that services inside the DSK
// backend subscribe to.
type Broker struct {
	*Subscribable

	// Incoming messages are sent here.
	incoming chan *Message

	// Quit channel, receiving true, when de-initialized.
	done chan bool
}

func (b *Broker) Open() error {
	go func() {
		for {
			select {
			case m := <-b.incoming:
				b.NotifyAll(m)
			case <-b.done:
				log.Print("Closing message broker (received quit)...")
				return
			}
		}
	}()
	return nil
}

func (b *Broker) Close() error {
	b.done <- true
	b.UnsubscribeAll()
	return nil
}

func (b *Broker) Accept(topic string, text string) bool {
	return b.AcceptMessage(NewMessage(topic, text))
}

// Accept a message for fan-out. Will never block. When the
// buffer is full the message will be discarded and not delivered.
func (b *Broker) AcceptMessage(m *Message) (ok bool) {
	log.Printf("Accepting %s...", m)

	select {
	case b.incoming <- m:
		ok = true
	default:
		log.Printf("Message buffer full, discarded: %s", m)
		ok = false
	}
	return
}

// Connect will pass a subscribable messages through into this broker.
func (b *Broker) Connect(o *Subscribable, ns string) chan bool {
	log.Printf("Connecting broker onto namespace %s...", ns)

	return o.SubscribeFuncWithMessage("*", func(m *Message) error {
		log.Printf("Receiving message from connected broker and pushing into namespace %s...", ns)
		b.Accept(fmt.Sprintf("%s.%s", ns, m.Topic), m.Text)
		return nil
	})
}
