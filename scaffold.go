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

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type (
	// Event is a Context event.
	Event struct {
		Info
		err error
	}
	// Scaffold is a constructor for Module(s).
	Scaffold struct {
		mort      Mortar
		modulesMu sync.RWMutex
		modules   map[string]*moduleWrapper
		ch        chan<- Event
		skip      Skipper
	}
)

// Err returns nil if a Module was loaded or an error if a problem occurred.
func (e Event) Err() error {
	return e.err
}

// New constructs Scaffold(ing) to apply Mortar on Stone from Module(s).
func New(mort Mortar, opt ...Option) *Scaffold {
	s := &Scaffold{
		mort:    mort,
		modules: make(map[string]*moduleWrapper),
		skip:    DefaultSkipper,
	}
	for _, fn := range opt {
		fn(s)
	}
	return s
}

// Load loads Module(s) using a Context.
func (s *Scaffold) Load(ctx context.Context, mod ...Module) error {
	// register Module(s)
	var registered []Info
	s.modulesMu.Lock()
	for _, m := range mod {
		info := m.Info()
		if s.skip(info) {
			continue
		}
		if w, ok := s.modules[info.String()]; !ok || !w.loaded {
			s.modules[info.String()] = &moduleWrapper{
				Module: m,
			}
			registered = append(registered, info)
		}
	}
	s.modulesMu.Unlock()
	c := newContext(ctx, s)
	// load registered Module(s)
	if err := c.Load(registered...); err != nil {
		return fmt.Errorf("scaffold failed to load, %w", err)
	}
	return nil
}

// Stat returns the Load runtime for a Module.
func (s *Scaffold) Stat(info Info) time.Duration {
	s.modulesMu.RLock()
	defer s.modulesMu.RUnlock()
	mod, ok := s.modules[info.String()]
	if !ok {
		return 0
	}
	return mod.runtime
}

// hook conveniently wraps Mortar.Hook.
func (s *Scaffold) hook(stone ...Stone) error {
	return s.mort.Hook(stone...)
}

// get returns a registered Module if it exists.
func (s *Scaffold) get(info Info) (Module, bool, bool) {
	s.modulesMu.RLock()
	defer s.modulesMu.RUnlock()
	mod, ok := s.modules[info.String()]
	if !ok {
		return nil, false, false
	}
	return mod, true, mod.loaded
}

// set updates a Module with provision metadata.
func (s *Scaffold) set(info Info, start time.Time) {
	s.modulesMu.Lock()
	defer s.modulesMu.Unlock()
	mod, ok := s.modules[info.String()]
	if ok {
		mod.loaded = true
		mod.runtime = time.Since(start)
	}
}

// depend appends dependencies by Info to Module.
func (s *Scaffold) depend(mod Module, info ...Info) bool {
	s.modulesMu.RLock()
	defer s.modulesMu.RUnlock()
	w, ok := s.modules[mod.Info().String()]
	if !ok {
		return false
	}
	w.dependsOn(info...)
	return true
}

// publish is a publisher utility for Event(s).
func (s *Scaffold) publish(e Event) {
	if s.ch != nil {
		s.ch <- e
	}
}

// Append couples Scaffold(ing), thus merging Module stats.
func (s *Scaffold) Append(scaffolding ...*Scaffold) {
	s.modulesMu.Lock()
	defer s.modulesMu.Unlock()
	for _, scaffold := range scaffolding {
		scaffold.modulesMu.RLock()
		for ref, mod := range scaffold.modules {
			if w, ok := s.modules[ref]; !ok || !w.loaded {
				s.modules[ref] = mod
			}
		}
		scaffold.modulesMu.RUnlock()
	}
}

// loaded lists all registered Module(s) that have been loaded.
func (s *Scaffold) loaded() (info []Info) {
	s.modulesMu.RLock()
	defer s.modulesMu.RUnlock()
	for _, mod := range s.modules {
		if mod.loaded {
			info = append(info, mod.Info())
		}
	}
	return
}
