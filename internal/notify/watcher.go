// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package notify

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/atelierdisko/dsk/internal/bus"
	core "github.com/rjeczalik/notify"
)

func NewWatcher(path string) (*Watcher, error) {
	log.Printf("Initializing watcher on %s...", path)

	w := &Watcher{
		Subscribable: &bus.Subscribable{},
		path:         path,
		// Make the channel buffered to ensure we do not block. Notify will drop
		// an event if the receiver is not able to keep up the sending pace.
		changes: make(chan core.EventInfo, 1),
		done:    make(chan bool),
	}
	return w, w.Open()
}

type Watcher struct {
	*bus.Subscribable

	// Path to watch for changes.
	path string

	// Changes to the directory tree are send here.
	changes chan core.EventInfo

	// Quit channel, receiving true, when we are closed.
	done chan bool
}

func (w *Watcher) String() string {
	return fmt.Sprintf("watcher (...%s)", w.path[len(w.path)-10:])
}

// Open watcher to look for changes below root. Will filter out changes
// to paths where a segment of it is hidden.
func (w *Watcher) Open() error {
	if err := core.Watch(w.path+"/...", w.changes, core.All); err != nil {
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

				if anyPathSegmentIsHidden(pp) {
					continue Outer
				}
				log.Printf("Change detected on: %s", p)
				// Security: do not reveal full path, basename is safe
				// and okay to drive notifications.
				w.NotifyAll(bus.NewMessage("changed", filepath.Base(p)))
			case <-w.done:
				log.Print("Stopping watcher (received quit)...")
				return
			}
		}
	}()
	return nil
}

func (w *Watcher) Close() error {
	log.Printf("Closing %s...", w)

	close(w.done)
	core.Stop(w.changes)
	w.UnsubscribeAll()
	return nil
}

// Checks if any of the path segments in the given relative or
// absolute path is hidden.
func anyPathSegmentIsHidden(path string) bool {
	for path != "/" && path != "." {
		b := filepath.Base(path)

		if strings.HasPrefix(b, ".") {
			return true
		}
		path = filepath.Dir(path)
	}
	return false
}
