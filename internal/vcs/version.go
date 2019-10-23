// Copyright 2019 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vcs

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/coreos/go-semver/semver"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

func NewVersionFromRef(ref *plumbing.Reference) *Version {
	var name string
	refn := ref.Name()

	if refn.IsBranch() {
		name = fmt.Sprintf("dev-%s", refn.Short())
	} else if refn.IsTag() {
		// Ensure we don't trim the leading v of tags that begin with other v-words.
		if regexp.MustCompile(`^v[0-9]+`).MatchString(refn.Short()) {
			name = strings.TrimLeft(refn.Short(), "v")
		} else {
			name = refn.Short()
		}
	} else {
		name = refn.Short()
	}

	v := &Version{
		Name:     name,
		Ref:      ref,
		isLive:   false,
		isTag:    refn.IsTag(),
		isBranch: refn.IsBranch(),
		parsed:   nil,
	}

	parsed, err := semver.NewVersion(name)
	if err == nil { // Being spec compliant is optional.
		v.parsed = parsed
	}

	return v
}

// AsLiveVersion converts a version into a "live" one.
func AsLiveVersion(v *Version) *Version {
	v.Name = "live"
	v.isLive = true

	return v
}

type Version struct {
	Name     string
	Ref      *plumbing.Reference
	isLive   bool
	isTag    bool
	isBranch bool
	parsed   *semver.Version
}

func (v *Version) String() string {
	return v.Name
}
