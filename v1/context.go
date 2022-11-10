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
	"github.com/pedregon/mason/internal/stack"
	"go.uber.org/fx"
	"sync"
	"time"
)

// Build initializes a new fx.App.
func Build(c *Context) *fx.App {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return fx.New(c.options...)
}

type (
	// Context is the Module context.
	Context struct {
		context.Context
		logger  Logger
		mu      sync.RWMutex
		options []fx.Option
		modules map[string]*moduleWrapper
		stack   *stack.Stack[Info]
	}
	// moduleWrapper wraps Module to track load status.
	moduleWrapper struct {
		Module
		loaded bool
		deps   []Info
	}
)

// NewContext creates a new Context.
func NewContext(opt ...Option) *Context {
	c := new(Context)
	c.Context = context.TODO()
	c.logger = nopLogger{}
	c.modules = make(map[string]*moduleWrapper)
	c.stack = new(stack.Stack[Info])
	for _, o := range opt {
		o(c)
	}
	return c
}

// Fx injects service dependencies for fx.New.
func (c *Context) Fx(opt ...fx.Option) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.options = append(c.options, opt...)
	return
}

// Graph returns the Module dependency graph.
func (c *Context) Graph() (deps []Dependency) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, mod := range c.modules {
		for _, dep := range mod.deps {
			deps = append(deps, Dependency{From: mod.Info(), To: dep})
		}
	}
	return
}

// SetContext sets the internal context.Context.
func (c *Context) SetContext(ctx context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Context = ctx
}

// Completed returns whether all Module(s) have been loaded.
func (c *Context) Completed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, mod := range c.modules {
		if mod.loaded == false {
			return false
		}
	}
	return true
}

// Load loads Module(s).
func (c *Context) Load(info ...Info) (err error) {
	for _, i := range info {
		c.mu.RLock()
		mod, exist := c.modules[i.String()]
		c.mu.RUnlock()
		if !exist {
			err = ErrInvalidModule
			c.logger.Error("failed to load", i, err)
			return
		}
		if !mod.loaded {
			if current, ok := c.stack.Peek(); ok && current.String() == i.String() {
				err = ErrSelfReferentialDependency
				c.logger.Error("failed to load", i, err)
				return
			}
			if c.stack.Has(i) {
				err = ErrCircularDependency
				c.logger.Error("failed to load", i, err)
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
			c.mu.Unlock()
			for {
				if err = c.stack.Err(); err != nil {
					c.logger.Error("failed to load", i, err)
					return
				}
				if c.stack.Size()-1 == index {
					c.logger.Info("loaded", i, KV{Key: "runtime", Value: time.Since(start).String()})
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
