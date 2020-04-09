// Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bus

import (
	"log"
	"math/rand"
	"path/filepath"
	"sync"
)

type Subscriber struct {
	receive chan<- *Message
	topic   string
}

func (s *Subscriber) Close() error {
	close(s.receive)
	return nil
}

// Subscribable is meant to be embedded by other structs, that
// are acting as a message/event bus.
type Subscribable struct {
	sync.RWMutex

	// A map of channels currently subscribed to changes. Once we
	// receive a message from the tree we fan it out to all. Once a
	// channel is detected to be closed, we remove it.
	subscribed map[int]*Subscriber
}

func (s *Subscribable) NotifyAll(m *Message) {
	s.RLock()
	defer s.RUnlock()

	for id, sub := range s.subscribed {
		matched, _ := filepath.Match(sub.topic, m.Topic)
		if !matched {
			continue
		}
		log.Printf("Notifying subscriber about %s...", m)

		select {
		case sub.receive <- m:
			// Subscriber received.
		default:
			log.Printf("Subscriber %d cannot receive, buffer full", id)
		}
	}
}

// Subscribe to a given topic. The topic may contain wildcard
// characters ("*"). Use "*" to subscribe to all messages.
func (s *Subscribable) Subscribe(topic string) (int, <-chan *Message) {
	log.Printf("Subscribing to topic %s...", topic)

	s.Lock()
	defer s.Unlock()

	id := rand.Int()
	ch := make(chan *Message, 10)

	if s.subscribed == nil {
		s.subscribed = make(map[int]*Subscriber, 0)
	}
	s.subscribed[id] = &Subscriber{receive: ch, topic: topic}
	return id, ch
}

// SubscribeFunc registers a handler func that is invoked on the
// given topic. Returns a quit chanel that can be used to stop the go
// routine, that runs the handler.
func (s *Subscribable) SubscribeFunc(topic string, fn func() error) chan bool {
	return s.SubscribeFuncWithMessage(topic, func(m *Message) error {
		return fn()
	})
}

func (s *Subscribable) SubscribeFuncWithMessage(topic string, fn func(*Message) error) chan bool {
	done := make(chan bool)
	go func() {
		id, messages := s.Subscribe(topic)

		for {
			select {
			case m, ok := <-messages:
				if !ok {
					log.Print("Stopping subscriber (channel closed)...")
					s.Unsubscribe(id)
					return
				}
				if err := fn(m); err != nil {
					log.Printf("Failed to run subscriber to %s: %s", topic, err)
				}
			case <-done:
				log.Print("Stopping subscriber (received quit)...")
				s.Unsubscribe(id)
				return
			}
		}
	}()
	return done
}

func (s *Subscribable) Unsubscribe(id int) {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.subscribed[id]; ok {
		s.subscribed[id].Close()
		delete(s.subscribed, id)
	}
}

func (s *Subscribable) UnsubscribeAll() {
	s.Lock()
	defer s.Unlock()

	for id := range s.subscribed {
		s.subscribed[id].Close()
		delete(s.subscribed, id)
	}
}
