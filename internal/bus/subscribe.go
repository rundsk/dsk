// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bus

import (
	"log"
	"math/rand"
	"sync"
)

type Subscriber struct {
	receive chan<- interface{}
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
		if sub.topic != "*" && sub.topic != m.Topic {
			continue
		}
		select {
		case sub.receive <- m:
			// Subscriber received.
		default:
			log.Printf("Subscriber %d cannot receive, buffer full", id)
		}
	}
}

func (s *Subscribable) Subscribe(topic string) (int, <-chan interface{}) {
	s.Lock()
	defer s.Unlock()

	id := rand.Int()
	ch := make(chan interface{}, 10)

	if s.subscribed == nil {
		s.subscribed = make(map[int]*Subscriber, 0)
	}
	s.subscribed[id] = &Subscriber{receive: ch, topic: topic}
	return id, ch
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
