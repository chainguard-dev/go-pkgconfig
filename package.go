// Copyright 2023 Chainguard, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pkgconfig

import (
	"os"
)

// Package contains the relevant information from a pkg-config file.
type Package struct {
	Name        string
	Description string
	Version     string
	URL         string
	Cflags      string
	Libs        string
	LibsPrivate string

	Requires         []Dependency
	RequiresPrivate  []Dependency
	RequiresInternal []Dependency
	Provides         []Dependency
}

type VersionCompare int

const (
	VersionEqual            = VersionCompare(0)
	VersionLessThan         = VersionCompare(-2)
	VersionLessThanEqual    = VersionCompare(-1)
	VersionGreaterThanEqual = VersionCompare(1)
	VersionGreaterThan      = VersionCompare(2)
)

// Dependency describes a dependency relationship between pkg-config files.
type Dependency struct {
	Identifier     string
	VersionCompare VersionCompare
	Version        string
}

// Parse parses a pkg-config data blob into a Package or returns an error.
func Parse(data []byte) (*Package, error) {
	return nil, nil
}

// Load loads a pkg-config data file from disk and returns a Package or an error.
func Load(path string) (*Package, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return Parse(data)
}
