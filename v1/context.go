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
	"sync"
	"time"
)

type (
	// Context is the context for Module(s).
	Context struct {
		context.Context
		logger  Logger
		mortar  Mortar
		mu      sync.RWMutex
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

// NewContext creates a new Context using Mortar.
func NewContext(mortar Mortar, opt ...Option) *Context {
	c := new(Context)
	c.Context = context.TODO()
	c.mortar = mortar
	c.modules = make(map[string]*moduleWrapper)
	c.stack = new(stack.Stack[Info])
	for _, o := range opt {
		o(c)
	}
	return c
}

// Hook injects service dependencies into Context.
func (c *Context) Hook(s ...Stone) error {
	return c.mortar.Hook(s...)
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
			if c.stack.Size() > 0 {
				err = ErrMissingDependency
			}
			if c.logger != nil {
				c.logger.Error("failed to load", i, err)
			}
			return
		}
		if !mod.loaded {
			if current, ok := c.stack.Peek(); ok && current.String() == i.String() {
				err = ErrSelfReferentialDependency
				if c.logger != nil {
					c.logger.Error("failed to load", i, err)
				}
				return
			}
			if c.stack.Has(i) {
				err = ErrCircularDependency
				if c.logger != nil {
					c.logger.Error("failed to load", i, err)
				}
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
					if c.logger != nil {
						c.logger.Error("failed to load", i, err)
					}
					return
				}
				if c.stack.Size()-1 == index {
					if c.logger != nil {
						c.logger.Info("loaded", i, KV{Key: "runtime", Value: time.Since(start).String()})
					}
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
