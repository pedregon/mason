// Copyright (c) 2023 pedregon
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

type (
	// Mortar is the "glue" for mounting some API. It represents a collection of loosely coupled application logic
	// intended to encourage inversion of control (IoC).
	// https://learn.microsoft.com/en-us/dotnet/architecture/modern-web-apps-azure/architectural-principles#dependency-inversion
	// It is recommended to use an interface guard.
	// https://caddyserver.com/docs/extending-caddy#interface-guards
	Mortar interface {
		// Hook mounts a Stone to some API.
		Hook(...Stone) error
	}
	// Stone is a "provider" that extends some API. Empty for future backwards compatibility.
	Stone any
)

// Len returns the number of Module(s) registered in Scaffold(ing).
func Len(s *Scaffold) (i int) {
	s.modulesMu.RLock()
	defer s.modulesMu.RUnlock()
	for range s.modules {
		i++
	}
	return
}

// Graph returns the Module dependency graph for Scaffold(ing).
func Graph(s *Scaffold) (deps []Dependency) {
	s.modulesMu.RLock()
	defer s.modulesMu.RUnlock()
	for _, mod := range s.modules {
		deps = append(deps, mod.listDeps()...)
	}
	return
}

// Loaded lists all Module(s) that have been loaded from underlying Scaffold(ing).
func Loaded(c *Context) (info []Info) {
	return c.scaffold.loaded()
}
