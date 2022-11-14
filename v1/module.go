// Copyright (c) 2022 miche.io
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
//
// SPDX-License-Identifier: MIT

package mason

import (
	"errors"
	"github.com/cespare/xxhash/v2"
)

var (
	ErrInvalidModule             error = errors.New("invalid module")
	ErrSelfReferentialDependency error = errors.New("self-referential module dependency")
	ErrCircularDependency        error = errors.New("circular module dependency")
	ErrMissingDependency         error = errors.New("missing module dependency")
)

type (
	// Info is Module information.
	Info struct {
		Name    string
		Version string
	}
	// Module is a compile-time plugin based on https://caddyserver.com/docs/extending-caddy.
	Module interface {
		// Info returns identifying information.
		Info() Info
		// Provision initializes.
		Provision(c *Context) error
	}
	// Dependency is a Module dependency relationship.
	Dependency struct {
		From Info
		To   Info
	}
)

// ID implements graph.Node from https://github.com/gonum/gonum.
func (i Info) ID() int64 {
	return int64(xxhash.Sum64String(i.String()))
}

// String implements fmt.Stringer.
func (i Info) String() string {
	return i.Name + "-" + i.Version
}

// String implements fmt.Stringer.
func (d Dependency) String() string {
	return d.To.String() + " <= " + d.From.String()
}
