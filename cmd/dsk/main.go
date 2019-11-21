// Copyright 2020 Marius Wilms. All rights reserved.
// Copyright 2017 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/rundsk/dsk/internal/api"

	"github.com/rundsk/dsk/internal/httputil"
	"github.com/rundsk/dsk/internal/plex"
	"github.com/fatih/color"
	isatty "github.com/mattn/go-isatty"
)

var (
	// Version string, compiled in.
	Version string

	// app is the global instance of the application.
	app *plex.App

	// sigc is the OS signal channel.
	sigc chan os.Signal
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
			if app != nil {
				log.Print("Cleaning up...")
				if err := app.Close(); err != nil {
					log.Printf("Failed to clean up: %s", err)
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
	start := time.Now()

	// This is nice to have, when requesting the output of "lsof -p
	// <PID>", when debugging unclosed file descriptors.
	log.Printf("Our PID: %d", os.Getpid())

	var livePath string
	if flag.Arg(0) != "" {
		livePath = flag.Arg(0)
	} else {
		// When no path is given as an argument, take the path to
		// the process itself. This makes sure that when opening the
		// binary from Finder the folder it is stored in is used.
		livePath = filepath.Dir(os.Args[0])
	}
	livePath, err := filepath.Abs(livePath)
	if err != nil {
		panic(err)
	}
	livePath, err = filepath.EvalSymlinks(livePath)
	if err != nil {
		panic(err)
	}
	log.Printf("Detected live path: %s", livePath)

	app = plex.NewApp( // assign to global
		Version,
		livePath,
		*ffrontend,
	)
	ctx, cancel := context.WithCancel(context.Background())
	app.Teardown.AddCancelFunc(cancel)

	if err := app.Open(); err != nil {
		log.Fatal(red.Sprintf("Failed to initialize application: %s", err))
	}

	if app.HasMultiVersionsSupport() {
		log.Printf("Detected support for multi-versions")

		if err := app.OpenVersions(ctx); err != nil {
			log.Print(red.Sprintf("Failed to start application: %s", err))
		}
	}

	apis := map[int]httputil.Mountable{
		1: api.NewV1(app.Sources, app.Version, app.Broker),
		2: api.NewV2(app.Sources, app.Version, app.Broker),
	}
	for _, a := range apis {
		a.MountHTTPHandlers()
	}
	// Must come last, as it contains a catch all route.
	app.Frontend.MountHTTPHandlers()

	addr := fmt.Sprintf("%s:%s", *host, *port)
	if isTerminal {
		log.Print("-------------------------------------------")
		log.Printf("Please visit: %s", green.Sprint("http://"+addr))
		log.Print("Hit Ctrl+C to quit")
		log.Print("-------------------------------------------")
	}
	log.Printf("Started web interface on %s, in %s", addr, time.Since(start))

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(red.Sprintf("Failed to start web interface: %s", err))
	}
}
