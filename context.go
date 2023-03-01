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
	"github.com/pedregon/mason/v2/internal/stack"
	"time"
)

var (
	// interface guard.
	_ Mortar = (*Context)(nil)
)

type (
	// Context is a context for loading Module(s) registered in a Scaffold.
	Context struct {
		context.Context
		scaffold *Scaffold
		stack    *stack.Stack[Info]
	}
)

// newContext creates a new Context for a Scaffold.
func newContext(ctx context.Context, scaffold *Scaffold) *Context {
	c := new(Context)
	c.Context = ctx
	c.scaffold = scaffold
	c.stack = new(stack.Stack[Info])
	return c
}

// Hook hooks Stone to mount points for Mortar.
func (c *Context) Hook(stone ...Stone) error {
	if err := c.Err(); err != nil {
		return err
	}
	return c.scaffold.hook(stone...)
}

// Load loads Module dependencies by Info.
func (c *Context) Load(info ...Info) (err error) {
	if err = c.Err(); err != nil {
		return
	}
	for _, i := range info {
		mod, exist, loaded := c.scaffold.get(i)
		if !exist {
			err = ErrInvalidModule
			if c.stack.Size() > 0 {
				err = ErrMissingDependency
			}
			c.scaffold.publish(Event{Info: i, err: err})
			return
		}
		if !loaded {
			if current, ok := c.stack.Peek(); ok && current.String() == i.String() {
				err = ErrSelfReferentialDependency
				c.scaffold.publish(Event{Info: i, err: err})
				return
			}
			if c.stack.Has(i) {
				err = ErrCircularDependency
				c.scaffold.publish(Event{Info: i, err: err})
				return
			}
			c.stack.Push(i)
			index := c.stack.Size() - 1
			start := time.Now()
			if err = mod.Provision(c); err != nil {
				c.stack.Log(err)
				return
			}
			c.scaffold.set(i, start)
			for {
				if err = c.stack.Err(); err != nil {
					c.scaffold.publish(Event{Info: i, err: err})
					return
				}
				if c.stack.Size()-1 == index {
					c.scaffold.publish(Event{Info: i, err: err})
					break
				}
				if top, ok := c.stack.Pop(); ok {
					c.scaffold.depend(mod, top)
				}
			}
		}
	}
	return
}
