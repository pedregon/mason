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
	"time"
)

var (
	// DefaultSkipper skips no Module(s).
	DefaultSkipper Skipper = func(_ Info) bool {
		return false
	}
)

type (
	// Option is a functional option for NewContext.
	// https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
	Option func(*Context)
	// Skipper is a callback function that decides whether to skip a Module.
	Skipper func(Info) bool
	// Observer is a callback function that observes a read-only Event channel.
	Observer func(*Context, <-chan Event)
)

// PanicOption panics if a Module is registered twice.
func PanicOption(c *Context) {
	c.isPanic = true
}

// SkipOption skips a Module on Load.
func SkipOption(skip Skipper) Option {
	return func(c *Context) {
		c.skip = skip
	}
}

// WatchOption watches Event(s). The Observer is protected from Module caller(s).
func WatchOption(obs Observer) Option {
	return func(c *Context) {
		obs(c, c.ch)
	}
}

// TimeoutOption is equivalent to context.WithTimeout.
func TimeoutOption(timeout time.Duration) Option {
	return func(c *Context) {
		c.Context, c.cancel = context.WithTimeout(c, timeout)
	}
}

// DeadlineOption is equivalent to context.WithDeadline.
func DeadlineOption(d time.Time) Option {
	return func(c *Context) {
		c.Context, c.cancel = context.WithDeadline(c, d)
	}
}

// ValueOption is equivalent to context.WithValue.
func ValueOption(key any, val any) Option {
	return func(c *Context) {
		c.Context = context.WithValue(c, key, val)
	}
}
