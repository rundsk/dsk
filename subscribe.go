// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"math/rand"
	"sync"
)

type Subscribable struct {
	sync.RWMutex

	// A map of channels currently subscribed to changes. Once we
	// receive a message from the tree we fan it out to all. Once a
	// channel is detected to be closed, we remove it.
	subscribed map[int]chan<- interface{}
}

func (s *Subscribable) NotifyAll(msg interface{}) {
	s.RLock()
	defer s.RUnlock()

	log.Printf("Notifying %d subscriber/s about %v...", msg, len(s.subscribed))
	for id, sub := range s.subscribed {
		select {
		case sub <- msg:
			// Subscriber received.
		default:
			log.Printf("Subscriber %d cannot receive, buffer full", id)
		}
	}
}

func (s *Subscribable) Subscribe() (int, <-chan interface{}) {
	s.Lock()
	defer s.Unlock()

	id := rand.Int()
	ch := make(chan interface{}, 10)

	if s.subscribed == nil {
		s.subscribed = make(map[int]chan<- interface{}, 0)
	}
	s.subscribed[id] = ch
	return id, ch
}

func (s *Subscribable) Unsubscribe(id int) {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.subscribed[id]; ok {
		close(s.subscribed[id])
		delete(s.subscribed, id)
	}
}

func (s *Subscribable) UnsubscribeAll() {
	s.Lock()
	defer s.Unlock()

	for id := range s.subscribed {
		close(s.subscribed[id])
		delete(s.subscribed, id)
	}
}
