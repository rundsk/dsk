// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"regexp"
	"strings"

	"github.com/rjeczalik/notify"
)

func NewWatcher(path string) *Watcher {
	return &Watcher{
		Subscribable: &Subscribable{},
		path:         path,
		// Make the channel buffered to ensure we do not block. Notify will drop
		// an event if the receiver is not able to keep up the sending pace.
		changes: make(chan notify.EventInfo, 1),
		done:    make(chan bool),
	}
}

type Watcher struct {
	*Subscribable

	// Path to watch for changes.
	path string

	// Changes to the directory tree are send here.
	changes chan notify.EventInfo

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
				w.NotifyAll(prettyPath(p))
			case <-w.done:
				log.Print("Watcher is closing...")
				return
			}
		}
	}()
	return nil
}

func (w *Watcher) Close() {
	w.UnsubscribeAll()
	w.done <- true
	notify.Stop(w.changes)
}
