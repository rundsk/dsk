// Copyright 2017 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ddt

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/atelierdisko/dsk/internal/author"
	"github.com/atelierdisko/dsk/internal/bus"
	"github.com/atelierdisko/dsk/internal/config"
	"github.com/atelierdisko/dsk/internal/vcs"
	"github.com/fatih/color"
)

var (
	ErrNodeTreeRootNotFound = errors.New("no tree root found")
)

// Tries to find root directory either by looking at args or the
// current working directory. This function needs the full path to the
// binary as a first argument and optionally an explicitly given path
// as the second argument.
func FindNodeTreeRoot(binary string, given string) (string, error) {
	var here string

	if given != "" {
		here = given
	} else {
		// When no path is given as an argument, take the path to
		// the process itself. This makes sure that when opening the
		// binary from Finder the folder it is stored in is used.
		here = filepath.Dir(binary)
	}
	here, err := filepath.Abs(here)
	if err != nil {
		return here, err
	}
	p, err := filepath.EvalSymlinks(here)
	if err != nil {
		return "", err
	}
	return p, nil
}

// NewNodeTree construct and partially initializes a NodeTree. Returns
// an unsynced tree from path; you must finalize initialization using
// Sync() or by calling Start().
func NewNodeTree(
	path string,
	cdb *config.DB,
	adb *author.DB,
	r *vcs.Repo,
	b *bus.Broker,
) *NodeTree {
	statModifiedNode := DefaultNodeModifiedStatter
	// Use a Repo for calculating the modified time. This is trying
	// to provide a better solution for situations where the modified
	// date on disk may not reflect the actual modification date. This
	// is the case when the DDT was checked out from Git during a
	// build process step.
	if r != nil {
		statModifiedNode = func(n *Node) (time.Time, error) {
			m, err := r.Modified(n.Path)

			// Fall back to default file system based retrieval if a
			// repository is not available. Also covers present but
			// uncommitted files.

			if err != nil && err != vcs.ErrNoData {
				return m, err
			}
			if m.IsZero() {
				// log.Printf("Falling back to default modified statter for node: %s", pathutil.Pretty(n.Path))
				return DefaultNodeModifiedStatter(n)
			}
			return m, nil
		}
	}

	statModifiedPath := DefaultPathModifiedStatter
	// Use a Repo for calculating the modified time. This is trying
	// to provide a better solution for situations where the modified
	// date on disk may not reflect the actual modification date. This
	// is the case when the DDT was checked out from Git during a
	// build process step.
	statModifiedPath = func(path string) (time.Time, error) {
		m, err := r.Modified(path)

		// Fall back to default file system based retrieval if a
		// repository is not available. Also covers present but
		// uncommitted files.

		if err != nil && err != vcs.ErrNoData {
			return m, err
		}
		if m.IsZero() {
			// log.Printf("Falling back to default modified statter for path: %s", pathutil.Pretty(path))
			return DefaultPathModifiedStatter(path)
		}
		return m, nil
	}

	return &NodeTree{
		Path:             path,
		configDB:         cdb,
		statModifiedNode: statModifiedNode,
		statModifiedPath: statModifiedPath,
		authorsDB:        adb,
		repo:             r,
		broker:           b,
	}
}

type NodeTree struct {
	// Ensures the tree is locked, when it is being synced, to
	// prevent reads in the middle of syncs.
	sync.RWMutex

	// The absolute root path of the tree.
	Path string

	// Maps node URL paths to nodes, for quick lookup.
	lookup map[string]*Node

	// Ordered slice of un-normalized node URLs.
	ordered []string

	// The root node and entry point to the acutal tree.
	Root *Node `json:"root"`

	configDB *config.DB

	statModifiedNode NodeModifiedStatter

	statModifiedPath PathModifiedStatter

	authorsDB *author.DB

	// Repository, if the tree is version controlled.
	repo *vcs.Repo

	// A place where we can send filtered messages to.
	broker *bus.Broker
}

// NodeGetter retrieves nodes from the tree, using the node's relative
// URL. When the node cannot be found ok will be false.
type NodeGetter func(url string) (ok bool, n *Node, err error)

// NodesGetter retrieves all nodes from the tree.
type NodesGetter func() []*Node

func (t *NodeTree) CalculateHash() (string, error) {
	t.RLock()
	defer t.RUnlock()
	return t.Root.CalculateHash()
}

// Sync recursively crawls the given root directory, constructing a
// tree of nodes. Will rebuild the entire tree on every sync. This
// makes the algorithm really simple - as we don't need to do branch
// selection - but also slow.
//
// Nodes that are discover but fail to finalize their initialization
// using Node.Load() will not be skipped but kept in tree in
// a semi-initialized way. So that the their children are not
// disconnected and no gaps exist in tree branches.
func (t *NodeTree) Sync() error {
	t.Lock()
	defer t.Unlock()

	start := time.Now()
	yellow := color.New(color.FgYellow)

	var nodes []*Node

	err := filepath.Walk(t.Path, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f.IsDir() {
			isRoot := filepath.Base(t.Path) == f.Name()

			if strings.HasPrefix(f.Name(), ".") && !isRoot {
				return filepath.SkipDir
			}
			n := NewNode(
				path,
				t.Path,
				t.configDB.Data().Project,
				t.statModifiedNode,
				t.statModifiedPath,
				t.authorsDB.GetByEmail,
			)

			if err := n.Load(); err != nil {
				log.Print(yellow.Sprint(err))
			}
			nodes = append(nodes, n)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("Failed to walk directory tree %s: %s", t.Path, err)
	}

	// In the second pass we're doing two things: add the children
	// to the nodes and build up the lookup tables, as we're already
	// iterating the nodes.
	lookup := make(map[string]*Node)
	ordered := make([]string, 0, len(nodes))

	for _, n := range nodes {
		lookup[n.LookupURL()] = n
		ordered = append(ordered, n.UnnormalizedURL())

		for _, sn := range nodes {
			if filepath.Dir(n.Path) == sn.Path {
				n.Parent = sn
				continue
			}
			if filepath.Dir(sn.Path) == n.Path {
				n.Children = append(n.Children, sn)
				continue
			}
		}
	}
	sort.Strings(ordered)

	// Swap late, in event of error we keep the previous state.
	t.lookup = lookup
	t.ordered = ordered
	t.Root = lookup[""]

	total := len(lookup)
	took := time.Since(start)

	defer t.broker.Accept(bus.NewMessage(
		"tree.synced", fmt.Sprintf("%d node/s in %s", total, took),
	))

	log.Printf("Synced tree with %d total node/s in %s", total, took)
	return nil
}

// Returns the neighboring previous and next nodes for the given
// current ddt. When current node is the last or first node, the
// behavior is not to wrap around.
//
// Determines the next node following the given current ddt. This
// may either be the first child of the given node, if there are none
// the sibling node and - walking up the tree - if there is none the
// parents sibling ddt. The algorithm for determing the previous
// node is analogous.
func (t *NodeTree) NeighborNodes(current *Node) (prev *Node, next *Node, err error) {
	t.RLock()
	defer t.RUnlock()

	var ok bool

	key := sort.SearchStrings(t.ordered, current.UnnormalizedURL())

	// SearchString returns the next unused key, if the given string
	// isn't found.
	if key == len(t.ordered) {
		return nil, nil, fmt.Errorf("No node with URL path '%s' in tree", current.URL())
	}

	// Be sure current node isn't the first ddt.
	if key != 0 {
		ok, prev, err = t.Get(normalizeNodeURL(t.ordered[key-1]))
		if !ok || err != nil {
			return prev, next, err
		}
	}

	// Check if current node isn't the last ddt.
	if key != len(t.ordered)-1 {
		ok, next, err = t.Get(normalizeNodeURL(t.ordered[key+1]))
		if !ok || err != nil {
			return prev, next, err
		}
	}
	return prev, next, err
}

// Returns the number of total nodes in the tree.
func (t *NodeTree) TotalNodes() uint16 {
	t.RLock()
	defer t.RUnlock()

	return uint16(len(t.lookup))
}

// Retrieves a node from the tree, performs a case-insensitive match.
func (t *NodeTree) Get(url string) (ok bool, n *Node, err error) {
	t.RLock()
	defer t.RUnlock()

	if n, ok := t.lookup[lookupNodeURL(url)]; ok {
		return ok, n, nil
	}
	return false, &Node{}, nil
}

// GetAll nodes as a flat slice.
func (t *NodeTree) GetAll() []*Node {
	t.RLock()
	defer t.RUnlock()

	ns := make([]*Node, 0, len(t.lookup))
	for _, n := range t.lookup {
		ns = append(ns, n)
	}
	return ns
}
