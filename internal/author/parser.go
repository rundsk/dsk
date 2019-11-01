// Copyright 2020 Marius Wilms. All rights reserved.
// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package author

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type ParseAuthor struct {
	Email string
	Name  string
}

// Parses the contents of an AUTHORS.txt file in mailmap format.
// Currently supports the simple syntax only.
//
// See: https://github.com/git/git/blob/master/Documentation/mailmap.txt
//
// Parse lines looking like this:
//   Proper Name <commit@email.xx>
//   # this is a comment
//   Proper Name <commit@email.xx> # inline comment
func parse(r io.Reader) ([]*ParseAuthor, error) {
	var parsed []*ParseAuthor

	lineScanner := bufio.NewScanner(r)

	for lineScanner.Scan() {
		line := strings.TrimSpace(lineScanner.Text())

		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		inlineComment := strings.Index(line, "#")

		beginMail := strings.Index(line, "<")
		if beginMail == -1 || (inlineComment != -1 && inlineComment < beginMail) {
			return parsed, fmt.Errorf("expected opening angle bracket in line '%s'", line)
		}
		endMail := strings.LastIndex(line, ">")
		if endMail == -1 || (inlineComment != -1 && inlineComment < endMail) {
			return parsed, fmt.Errorf("expected closing angle bracket in line '%s'", line)
		}

		name := strings.TrimSpace(line[0 : beginMail-1])
		email := strings.TrimSpace(line[beginMail+1 : len(line)-1])
		parsed = append(parsed, &ParseAuthor{email, name})
	}
	return parsed, nil
}
