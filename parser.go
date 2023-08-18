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
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/ijt/goparsify"
)

type Property struct {
	Key   string
	Value string
}

type Variable struct {
	Key   string
	Value string
}

type DependencyList struct {
	Key          string
	Dependencies []Dependency
}

var (
	comment = goparsify.Seq("#", goparsify.Maybe(goparsify.NotChars("#\n")))

	key                = goparsify.Chars("a-zA-Z0-9_.")
	value              = goparsify.NotChars("\n").Map(func(n *goparsify.Result) { n.Result = n.Token })
	variableAssignment = goparsify.Seq(key, "=", goparsify.Maybe(value)).Map(func(n *goparsify.Result) {
		n.Result = Variable{
			Key:   n.Child[0].Token,
			Value: n.Child[2].Token,
		}
	})
	propertyAssignment = goparsify.Seq(key, goparsify.Cut(), ":", goparsify.Maybe(value)).Map(func(n *goparsify.Result) {
		key := strings.ToTitle(n.Child[0].Token)
		value := n.Child[3].Token

		n.Result = Property{
			Key:   key,
			Value: value,
		}
	})

	identifier = goparsify.Chars("a-zA-Z0-9-_.")
	version    = goparsify.Chars("a-zA-Z0-9-.")

	verLessThan         = goparsify.Exact("<").Map(func(n *goparsify.Result) { n.Result = VersionLessThan })
	verLessThanEqual    = goparsify.Exact("<=").Map(func(n *goparsify.Result) { n.Result = VersionLessThanEqual })
	verEqual            = goparsify.Exact("=").Map(func(n *goparsify.Result) { n.Result = VersionEqual })
	verGreaterThanEqual = goparsify.Exact(">").Map(func(n *goparsify.Result) { n.Result = VersionGreaterThanEqual })
	verGreaterThan      = goparsify.Exact(">=").Map(func(n *goparsify.Result) { n.Result = VersionGreaterThan })
	verMatch            = goparsify.Any(verLessThanEqual, verLessThan, verEqual, verGreaterThan, verGreaterThanEqual)

	dependencyChain = goparsify.Seq(identifier, verMatch, version)
	dependency      = goparsify.Any(dependencyChain, identifier)
	dependencyList  = goparsify.Many(dependency, goparsify.Maybe(","))

	dependencyListKeywords   = goparsify.Any("Requires.private", "Requires.internal", "Requires", "Provides")
	dependencyListAssignment = goparsify.Seq(dependencyListKeywords, ":", dependencyList).Map(func(n *goparsify.Result) {
		deps := []Dependency{}
		key := strings.ToTitle(n.Child[0].Token)

		for _, d := range n.Child[2].Child {
			vercmp, ok := d.Child[1].Result.(VersionCompare)
			if ok {
				dep := Dependency{
					Identifier:     d.Child[0].Token,
					VersionCompare: vercmp,
					Version:        d.Child[2].Token,
				}

				deps = append(deps, dep)
			} else {
				dep := Dependency{
					Identifier: d.Child[0].Token,
				}

				deps = append(deps, dep)
			}
		}

		n.Result = DependencyList{
			Key:          key,
			Dependencies: deps,
		}
	})

	documentChain = goparsify.Many(goparsify.Any("\n", comment, variableAssignment, dependencyListAssignment, propertyAssignment)).Map(func(n *goparsify.Result) {
		res := []interface{}{}

		for _, child := range n.Child {
			res = append(res, child.Result)
		}

		n.Result = res
	})
)

func matchWhitespace(s *goparsify.State) {
	for s.Pos < len(s.Input) {
		r, w := utf8.DecodeRuneInString(s.Get())
		if r == '\n' || !unicode.IsSpace(r) {
			return
		}
		s.Pos += w
	}
}

//nolint:unparam
func (pkg *Package) replaceVariables(input string) (string, error) {
	mutatedInput := input

	for k, v := range pkg.Vars {
		mutatedInput = strings.ReplaceAll(mutatedInput, fmt.Sprintf("${%s}", k), v)
	}

	return mutatedInput, nil
}

// Parse parses a pkg-config data blob into a Package or returns an error.
func Parse(data string) (*Package, error) {
	pkg := Package{Vars: map[string]string{}}

	result, _, err := goparsify.Run(documentChain, data, matchWhitespace)
	if err != nil {
		return nil, err
	}
	astTree, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("parse result is not AST")
	}

	for _, astNode := range astTree {
		switch specializedNode := astNode.(type) {
		case Variable:
			finalValue, err := pkg.replaceVariables(specializedNode.Value)
			if err != nil {
				return nil, err
			}

			pkg.Vars[specializedNode.Key] = finalValue

		case Property:
			finalValue, err := pkg.replaceVariables(specializedNode.Value)
			if err != nil {
				return nil, err
			}

			switch specializedNode.Key {
			case "NAME":
				pkg.Name = finalValue
			case "VERSION":
				pkg.Version = finalValue
			case "DESCRIPTION":
				pkg.Description = finalValue
			case "URL":
				pkg.URL = finalValue
			case "CFLAGS":
				pkg.Cflags = finalValue
			case "CFLAGS.PRIVATE":
				pkg.CflagsPrivate = finalValue
			case "LIBS":
				pkg.Libs = finalValue
			case "LIBS.PRIVATE":
				pkg.LibsPrivate = finalValue
			}

		case DependencyList:
			switch specializedNode.Key {
			case "REQUIRES":
				pkg.Requires = specializedNode.Dependencies
			case "REQUIRES.PRIVATE":
				pkg.RequiresPrivate = specializedNode.Dependencies
			case "REQUIRES.INTERNAL":
				pkg.RequiresInternal = specializedNode.Dependencies
			case "PROVIDES":
				pkg.Provides = specializedNode.Dependencies
			}
		}
	}

	return &pkg, nil
}
