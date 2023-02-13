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
)

var (
	ErrInvalidModule             error = errors.New("invalid module")
	ErrSelfReferentialDependency error = errors.New("self-referential module dependency")
	ErrCircularDependency        error = errors.New("circular module dependency")
	ErrMissingDependency         error = errors.New("missing module dependency")
	ErrDuplicateModule           error = errors.New("register called twice for module")
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

// String implements fmt.Stringer.
func (i Info) String() string {
	return i.Name + "-" + i.Version
}

// String implements fmt.Stringer.
func (d Dependency) String() string {
	return d.To.String() + " <= " + d.From.String()
}

// Graph returns the Module dependency graph for a Context.
func Graph(c *Context) (deps []Dependency) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, mod := range c.modules {
		for _, dep := range mod.deps {
			deps = append(deps, Dependency{From: mod.Info(), To: dep})
		}
	}
	return
}

// Load loads all Module(s) from a Context.
func Load(c *Context) error {
	c.mu.RLock()
	var info []Info
	for _, mod := range c.modules {
		info = append(info, mod.Info())
	}
	c.mu.RUnlock()
	return c.Load(info...)
}

// Loaded returns all Module(s) from a Context that have been loaded.
func Loaded(c *Context) (info []Info) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, mod := range c.modules {
		if mod.loaded {
			info = append(info, mod.Info())
		}
	}
	return
}

// Len returns the number of Module(s) registered in a Context.
func Len(c *Context) (i int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for range c.modules {
		i++
	}
	return
}
