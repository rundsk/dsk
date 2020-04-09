// Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
// Copyright 2019 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vcs

type Versions struct {
	data []*Version
}

func (vs *Versions) Len() int {
	return len(vs.data)
}

func (vs *Versions) Swap(i, j int) {
	vs.data[i], vs.data[j] = vs.data[j], vs.data[i]
}

// Less will help sort on a best effort basis.
//
// First names of tags and names of branches are separately clustered.
// Tags come before branches. But the special "live" version comes
// always first.
//
// Inside the tag cluster tags following semver are sorted according
// to its rules. If they do not follow the spec they are sorted
// lexicographically.
//
// Inside the branch cluster names are always sorted lexicographically.
func (vs *Versions) Less(i, j int) bool {
	if vs.data[i].isLive {
		return false
	}
	if vs.data[j].isLive {
		return true
	}
	if vs.data[i].isBranch && vs.data[j].isTag {
		return true
	}
	if vs.data[i].isTag && vs.data[j].isBranch {
		return false
	}

	if vs.data[i].isBranch && vs.data[j].isBranch {
		return len(vs.data[i].Name) < len(vs.data[j].Name)
	}

	if vs.data[i].isTag && vs.data[j].isTag {
		if vs.data[i].parsed == nil || vs.data[j].parsed == nil {
			return len(vs.data[i].Name) < len(vs.data[j].Name)
		}
		return vs.data[i].parsed.LessThan(*vs.data[j].parsed)
	}

	return len(vs.data[i].Name) < len(vs.data[j].Name)
}

func (vs *Versions) Add(v *Version) {
	vs.data = append(vs.data, v)
}

func (vs *Versions) ForEach(fn func(*Version) error) error {
	for _, v := range vs.data {
		if err := fn(v); err != nil {
			return err
		}
	}
	return nil
}

func (vs *Versions) Filter(fn func(*Version) bool) *Versions {
	filtered := make([]*Version, 0, len(vs.data))

	for _, v := range vs.data {
		if fn(v) {
			filtered = append(filtered, v)
		}
	}
	return &Versions{filtered}
}

func (vs *Versions) Names() []string {
	names := make([]string, 0, len(vs.data))

	for _, v := range vs.data {
		names = append(names, v.Name)
	}
	return names
}
