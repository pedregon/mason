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

type (
	// Option is a functional option for NewContext.
	// https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
	Option func(*Context)
)

// ModuleOption registers Module(s) that may be loaded.
func ModuleOption(mod ...Module) Option {
	return func(c *Context) {
		for _, m := range mod {
			info := m.Info()
			w := new(moduleWrapper)
			w.Module = m
			c.modules[info.String()] = w
		}
	}
}

// LoggerOption overwrites the default Logger.
func LoggerOption(logger Logger) Option {
	return func(c *Context) {
		c.logger = logger
	}
}
