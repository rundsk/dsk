// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build tools

package main

import (
	"log"
	"net/http"
	"os"

	"github.com/shurcooL/vfsgen"
)

func main() {
	env, ok := os.LookupEnv("FRONTEND")
	var fs http.FileSystem
	if !ok {
		log.Print("No FRONTEND environment variable found using: ./frontend ")
		fs = http.Dir("./frontend")
	} else {
		log.Printf("Using FRONTEND environment variable: %s", env)
		fs = http.Dir(env)
	}

	err := vfsgen.Generate(fs, vfsgen.Options{
		Filename:  "frontend_vfsdata.go",
		BuildTags: "!dev",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
