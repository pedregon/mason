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
	"context"
	"fmt"
	"github.com/pedregon/mason/internal/stack"
	"sync"
	"time"
)

var (
	// interface guard.
	_ Mortar = (*Context)(nil)
)

type (
	// Event is a Context event.
	Event struct {
		info Info
		err  error
	}
	// Context is the context for Module(s).
	Context struct {
		context.Context
		mort     Mortar
		mu       sync.RWMutex
		modules  map[string]*moduleWrapper
		stack    *stack.Stack[Info]
		ch       chan Event
		isClosed bool
		skip     Skipper
		isPanic  bool
		cancel   context.CancelFunc
	}
	// moduleWrapper wraps Module to track load status.
	moduleWrapper struct {
		Module
		loaded  bool
		runtime time.Duration
		deps    []Info
	}
)

// Blame returns the Module responsible.
func (e Event) Blame() Info {
	return e.info
}

// Err returns nil if a Module was loaded or an error if a problem occurred.
func (e Event) Err() error {
	return e.err
}

// NewContext creates a new Context using Mortar.
func NewContext(ctx context.Context, mort Mortar, opt ...Option) (*Context, context.CancelFunc) {
	c := new(Context)
	c.mort = mort
	c.modules = make(map[string]*moduleWrapper)
	c.stack = new(stack.Stack[Info])
	c.ch = make(chan Event, 1)
	c.skip = DefaultSkipper
	c.Context, c.cancel = context.WithCancel(ctx)
	for _, o := range opt {
		o(c)
	}
	return c, func() {
		c.cancel()
		if !c.isClosed {
			close(c.ch)
		}
		c.isClosed = true
	}
}

// Register registers a Module to the Context.
func (c *Context) Register(mod ...Module) (err error) {
	if err = c.Err(); err != nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, m := range mod {
		info := m.Info()
		if m == nil {
			err = ErrInvalidModule
			if c.isPanic {
				panic(err)
			}
			c.ch <- Event{info: info, err: err}
			return
		}
		if _, dup := c.modules[info.String()]; dup {
			err = fmt.Errorf("%w %s", ErrDuplicateModule, info.String())
			if c.isPanic {
				panic(err)
			}
			c.ch <- Event{info: info, err: err}
			return
		}
		c.modules[info.String()] = &moduleWrapper{
			Module: m,
		}
	}
	return
}

// Hook injects service dependencies into Context.
func (c *Context) Hook(s ...Stone) error {
	if err := c.Err(); err != nil {
		return err
	}
	return c.mort.Hook(s...)
}

// Load loads Module(s).
func (c *Context) Load(info ...Info) (err error) {
	if err = c.Err(); err != nil {
		return
	}
	for _, i := range info {
		c.mu.RLock()
		mod, exist := c.modules[i.String()]
		c.mu.RUnlock()
		if !exist {
			err = ErrInvalidModule
			if c.stack.Size() > 0 {
				err = ErrMissingDependency
			}
			c.ch <- Event{info: i, err: err}
			return
		}
		if c.skip(i) {
			continue
		}
		if !mod.loaded {
			if current, ok := c.stack.Peek(); ok && current.String() == i.String() {
				err = ErrSelfReferentialDependency
				c.ch <- Event{info: i, err: err}
				return
			}
			if c.stack.Has(i) {
				err = ErrCircularDependency
				c.ch <- Event{info: i, err: err}
				return
			}
			c.stack.Push(i)
			index := c.stack.Size() - 1
			start := time.Now()
			if err = mod.Provision(c); err != nil {
				c.stack.Log(err)
				return
			}
			c.mu.Lock()
			mod.loaded = true
			mod.runtime = time.Since(start)
			c.mu.Unlock()
			for {
				if err = c.stack.Err(); err != nil {
					c.ch <- Event{info: i, err: err}
					return
				}
				if c.stack.Size()-1 == index {
					c.ch <- Event{info: i, err: err}
					break
				}
				if last, ok := c.stack.Pop(); ok {
					c.mu.Lock()
					mod.deps = append(mod.deps, last)
					c.mu.Unlock()
				}
			}
		}
	}
	return
}

// Stat returns the Load runtime for a Module.
func (c *Context) Stat(info Info) time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	mod, ok := c.modules[info.String()]
	if !ok {
		return 0
	}
	return mod.runtime
}
