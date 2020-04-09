// Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
// Copyright 2019 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vcs

import (
	"fmt"
	"log"
	"sync"
	"time"

	"gopkg.in/src-d/go-git.v4/plumbing"
)

type LookupBuildFn func() (*plumbing.Reference, interface{}, error)

func NewLookup(scope string, fn LookupBuildFn) (*Lookup, error) {
	l := &Lookup{
		scope:      scope,
		buildFn:    fn,
		buildReqs:  make(chan *BuildRequest, 5000),
		builds:     make(chan *BuildRequest),
		builtQueue: make(chan *BuildRequest, 5000),
		done:       make(chan bool),
	}
	return l, l.Open()
}

type BuildRequest struct {
	ref *plumbing.Reference
	ch  chan interface{}
}

func (br *BuildRequest) String() string {
	return fmt.Sprintf("build request (%s)", br.ref.Name().Short())
}

type Lookup struct {
	sync.RWMutex

	// scope identifies what this lookup is for.
	scope string

	ref *plumbing.Reference

	data interface{}

	buildReqs chan *BuildRequest

	builds chan *BuildRequest

	// builtQueue enqueues requests that can be satisfied once the
	// builcurrent d has finished.
	builtQueue chan *BuildRequest

	// buildFn is a func that actually performs the build and returns
	// the result.
	buildFn LookupBuildFn

	// buildingRef is the reference currently being built. When we
	// receive a request for building the same ref, we can notify it
	// when the currently build ref is done and discard the original
	// request.
	buildingRef *plumbing.Reference

	done chan bool
}

func (l *Lookup) String() string {
	return fmt.Sprintf("lookup table (%s)", l.scope)
}

// Open starts a goroutine to hande build requests.
func (l *Lookup) Open() error {
	go func() {
		for {
			select {
			case req := <-l.builds:
				log.Printf("Received %s in %s, running build...", req, l)

				l.Lock()
				l.buildingRef = req.ref
				l.Unlock()

				log.Printf("Building %s...", l)
				start := time.Now()

				bref, bl, berr := l.buildFn()
				if berr != nil {
					continue
				}
				log.Printf("Built %s in %s", l, time.Since(start))

				l.Lock()
				l.buildingRef = nil
				l.ref = bref
				l.data = bl
				l.Unlock()

				log.Printf("Answering original %s in %s...", req, l)
				req.ch <- bl // Notify the original request, that triggered the build.

				if len(l.builtQueue) > 0 {
					log.Printf("Answering %d queued up build requests in %s...", len(l.builtQueue), l)
				}
				for len(l.builtQueue) > 0 {
					qreq := <-l.builtQueue
					qreq.ch <- bl
				}
				log.Printf("Handled %s in %s.", req, l)
			case <-l.done:
				log.Printf("Stopping %s build request handler (received quit)...", l)
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case req := <-l.buildReqs:
				if req.ref == nil { // Invalid request, cannot satisfy.
					log.Printf("Got build request for nil ref, queuing for next build in %s.", l)
					l.builtQueue <- req
					continue
				}

				l.RLock()
				lref := l.ref
				lbref := l.buildingRef
				l.RUnlock()

				log.Printf("Dispatching %s in %s lookup...", req, l.scope)
				if lref == nil { // No build happened, run one.
					log.Print("No build happened, running one.")
					l.builds <- req
				} else if lref.Hash() == req.ref.Hash() { // Current request matches current state.
					log.Print("Current request matches current state.")
					req.ch <- l.data
				} else if lbref == nil { // No build running, request cannot be satisified with current state.
					log.Print("No build running, request cannot be satisified with current state.")
					l.builds <- req
				} else if lbref.Hash() == req.ref.Hash() { // Build for same ref running.
					log.Print("Build for same ref running.")
					l.builtQueue <- req
				} else {
					log.Print("How did we get here?")
					l.builds <- req
				}
				log.Printf("Dispatched %s in %s lookup.", req, l.scope)
			case <-l.done:
				log.Printf("Stopping %s lookup build request dispatcher (received quit)...", l.scope)
				return
			}
		}
	}()

	return nil
}

func (l *Lookup) Close() error {
	log.Printf("Closing %s...", l)
	l.done <- true
	return nil
}

func (l *Lookup) RequestBuild(ref *plumbing.Reference) {
	req := &BuildRequest{ref, make(chan interface{}, 1)}
	l.buildReqs <- req
}

// GetDirtyOkay returns the lookup table even when it doesn't match
// the given ref. However the lookup table must have been build at
// least once.
//
// If this isn't the case a build will be triggered and the caller
// can wait until it succeeds. When nil is requested will queue the
// request and answer it once any ref of the lookup has been built.
func (l *Lookup) GetDirtyOkay(ref *plumbing.Reference) chan interface{} {
	l.RLock()
	lref := l.ref
	ldata := l.data
	l.RUnlock()

	ch := make(chan interface{}, 1)

	if lref != nil {
		ch <- ldata
		return ch
	}
	req := &BuildRequest{ref, ch}
	l.buildReqs <- req
	return req.ch
}

func (l *Lookup) IsStale(head *plumbing.Reference) bool {
	if head == nil {
		return false
	}
	if l.ref == nil {
		return true
	}
	return l.ref.Hash() != head.Hash()
}
