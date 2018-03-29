// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"math/rand"
	"regexp"
	"strings"
	"sync"

	"github.com/rjeczalik/notify"
)

func NewWatcher(path string) *Watcher {
	return &Watcher{
		path: path,
		// Make the channel buffered to ensure we do not block. Notify will drop
		// an event if the receiver is not able to keep up the sending pace.
		changes:    make(chan notify.EventInfo, 1),
		subscribed: make(map[int]chan<- string, 0),
		done:       make(chan bool),
	}
}

type Watcher struct {
	// Protects the subscriber map.
	sync.RWMutex

	// Path to watch for changes.
	path string

	// Changes to the directory tree are send here.
	changes chan notify.EventInfo

	// A map of channels currently subscribed to changes. Once
	// we receive a message from the tree we fan it out to all. Once a
	// channel is detected to be closed, we remove it from here.
	subscribed map[int]chan<- string

	// Quit channel, receiving true, when we are closed.
	done chan bool
}

// Open watcher to look for changes below root. Will filter out changes
// to paths where a segment of it matches the ignore regexp.
func (w *Watcher) Open(ignore *regexp.Regexp) error {
	if err := notify.Watch(w.path+"/...", w.changes, notify.All); err != nil {
		return err
	}

	go func() {
	Outer:
		for {
			select {
			case ei := <-w.changes:
				p := ei.Path()

				// Do not match directories above tree root. If we
				// are placed inside an ignored dir, everything will
				// always be ignored. Even if the tree root directory
				// is set to be ignored, do not ignore it, as the tree
				// has been intentionally loaded from that directory.
				pp := strings.TrimPrefix(p, w.path+"/")

				if anyPathSegmentMatches(pp, ignore) {
					continue Outer
				}
				log.Printf("Watcher detected change on: %s", prettyPath(p))

				w.RLock()
				if len(w.subscribed) != 0 {
					log.Printf("Notifying %d watch subscriber/s...", len(w.subscribed))
				}
				for id, sub := range w.subscribed {
					select {
					case sub <- p:
						// Subscriber received.
					default:
						log.Printf("Watch subscriber %d cannot receive, buffer full", id)
					}
				}
				w.RUnlock()
			case <-w.done:
				log.Print("Watcher is closing...")
				return
			}
		}
	}()
	return nil
}

func (w *Watcher) Close() {
	w.Lock()

	for id := range w.subscribed {
		close(w.subscribed[id])
		delete(w.subscribed, id)
	}
	w.Unlock()

	w.done <- true
	notify.Stop(w.changes)
}

func (w *Watcher) Subscribe() (int, <-chan string) {
	w.Lock()
	defer w.Unlock()

	id := rand.Int()
	ch := make(chan string, 10)

	w.subscribed[id] = ch
	return id, ch
}

func (w *Watcher) Unsubscribe(id int) {
	w.Lock()
	defer w.Unlock()

	if _, ok := w.subscribed[id]; ok {
		close(w.subscribed[id])
		delete(w.subscribed, id)
	}
}
