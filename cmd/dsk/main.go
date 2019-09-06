// Copyright 2017 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/atelierdisko/dsk/internal/api"
	"github.com/atelierdisko/dsk/internal/ddt"
	"github.com/atelierdisko/dsk/internal/httputil"
	"github.com/atelierdisko/dsk/internal/pathutil"
	"github.com/fatih/color"
	isatty "github.com/mattn/go-isatty"
)

type CleanupFunc func() error

var (
	// Version string, compiled in.
	Version string

	// OS Signal channel.
	sigc chan os.Signal

	// teardown stacks functions for teardown, when the program exits.
	// Otherwise we'd need to keep the services we want to stop around
	// in globals.
	teardown []CleanupFunc
)

func main() {
	// Disable prefix, we are invoked directly.
	log.SetFlags(0)
	isTerminal := isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())

	// Listen for interrupt and allow to cancel program early.
	sigc = make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt)
	go func() {
		for sig := range sigc {
			log.Printf("Caught %v signal, bye!", sig)
			log.Print("Cleaning up...")

			// Close services in reverse order of starting them.
			for i := len(teardown) - 1; i >= 0; i-- {
				if teardown[i] == nil {
					continue
				}
				err := teardown[i]()

				if err != nil {
					log.Printf("Failed to teardown: %s", err)
				}
			}
			os.Exit(1)
		}
	}()

	host := flag.String("host", "127.0.0.1", "host IP to bind to")
	port := flag.String("port", "8080", "port to bind to")
	version := flag.Bool("version", false, "print DSK version")
	noColor := flag.Bool("no-color", false, "disables color output")
	ffrontend := flag.String("frontend", "", "path to a frontend, to use instead of the built-in")
	flag.Parse()

	if len(flag.Args()) > 1 {
		log.Fatalf("Too many arguments given, expecting exactly 0 or 1")
	}

	if *version {
		fmt.Println(Version)
		os.Exit(1)
	}

	// Color package automatically disables colors when not a TTY. We
	// don't need to check for an interactive terminal here again.
	if *noColor {
		color.NoColor = true
	}
	whiteOnBlue := color.New(color.FgWhite, color.BgBlue)
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)

	if isTerminal {
		log.Print("-------------------------------------------")
		log.Print(whiteOnBlue.Sprint(" DSK "))
		log.Printf("Version %s", Version)
		log.Print("-------------------------------------------")
	}

	// This is nice to have, when requesting the output of "lsof -p
	// <PID>", when debugging unclosed file descriptors.
	log.Printf("Our PID: %d", os.Getpid())

	here, err := ddt.FindNodeTreeRoot(os.Args[0], flag.Arg(0))
	if err != nil {
		log.Fatal(red.Sprintf("Failed to detect root of design definitions tree: %s", err))
	}

	log.Printf("Tree root found: %s", here)

	pathutil.SetPrettyRoot(here)

	broker, cleanup, err := initBroker()
	teardown = append(teardown, cleanup)
	if err != nil {
		log.Fatal(red.Sprintf("Failed to initialize message broker: %s", err))
	}

	watcher, cleanup, err := initWatcher(here)
	teardown = append(teardown, watcher.Stop)
	if err != nil {
		log.Fatal(red.Sprintf("Failed to initialize watcher: %s", err))
	}

	configDB, cleanup, err := initConfigDB(here, broker)
	teardown = append(teardown, configDB.Close)
	if err != nil {
		log.Fatal(red.Sprintf("Failed to initialize configuration database: %s", err))
	}

	repo, cleanup, err := initRepo(here, configDB)
	teardown = append(teardown, cleanup)
	if err != nil {
		log.Fatal(red.Sprintf("Failed to initialize repository: %s", err))
	}

	authorDB, cleanup, err := initAuthorDB(here, broker)
	teardown = append(teardown, cleanup)
	if err != nil {
		log.Fatal(red.Sprintf("Failed to initialize author database: %s", err))
	}

	tree, cleanup, err := initNodeTree(here, configDB, authorDB, repo, watcher, broker)
	teardown = append(teardown, cleanup)
	if err != nil {
		log.Fatal(red.Sprintf("Failed to initialize node tree: %s", err))
	}

	search, cleanup, err := initSearch(tree, broker, configDB)
	teardown = append(teardown, cleanup)
	if err != nil {
		log.Fatal(red.Sprintf("Failed to initialize search: %s", err))
	}

	frontend, cleanup, err := initFrontend(*ffrontend, tree)
	teardown = append(teardown, cleanup)
	if err != nil {
		log.Fatal(red.Sprintf("Failed to initialize frontend: %s", err))
	}

	apis := map[int]httputil.Mountable{
		1: api.NewV1(configDB, Version, tree, broker, search),
		2: api.NewV2(configDB, Version, tree, broker, search),
	}
	for v, a := range apis {
		a.MountHTTPHandlers()
		log.Printf("Mounted HTTP handlers: APIv%d", v)
	}

	// Must come last, as it contains a catch all route.
	frontend.MountHTTPHandlers()
	log.Print("Mounted HTTP handlers: frontend")

	addr := fmt.Sprintf("%s:%s", *host, *port)
	if isTerminal {
		log.Print("-------------------------------------------")
		log.Printf("Please visit: %s", green.Sprint("http://"+addr))
		log.Print("Hit Ctrl+C to quit")
		log.Print("-------------------------------------------")
	}
	log.Printf("Started web interface on %s", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(red.Sprintf("Failed to start web interface: %s", err))
	}
}
