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
	"errors"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/fx/fxtest"
	"strings"
	"testing"
	"time"
)

type (
	module struct {
		name     string
		version  string
		deps     []Info
		services []Service
	}
)

func (mod module) Info() (info Info) {
	info.Name = mod.name
	info.Version = mod.version
	return
}

func (mod module) Provision(c *Context) (err error) {
	if err = c.Load(mod.deps...); err != nil {
		return
	}
	if len(mod.services) > 0 {
		c.Hook(mod.services...)
	}
	return
}

type (
	testLogger struct {
		logger *testing.T
	}
)

func (t testLogger) Info(msg string, info Info, kv ...KV) {
	format := "[Modules] INFO msg=%s module=%s"
	for _, pair := range kv {
		format += " " + pair.String()
	}
	t.logger.Logf(format, msg, info)
}

func (t testLogger) Error(msg string, info Info, err error) {
	t.logger.Logf("[Modules] ERROR msg=%s info=%s err='%s'", msg, info, err.Error())
}

func TestModules(t *testing.T) {
	baz := &module{name: "baz", version: "1.0.0"}
	bar := &module{name: "bar", version: "1.0.0"}
	bar.deps = append(bar.deps, baz.Info())
	foo := &module{name: "foo", version: "1.0.0"}
	foo.deps = append(foo.deps, bar.Info())
	c := NewContext(ModuleOption(foo, bar, baz), LoggerOption(testLogger{logger: t}))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	c.SetContext(ctx)
	var err error
	go func() {
		err = c.Load(bar.Info(), foo.Info(), baz.Info())
		cancel()
	}()
	<-c.Done()
	if errors.Is(c.Err(), context.DeadlineExceeded) {
		t.Fatal(c.Err())
	}
	if err != nil {
		t.Fatal(err)
	}
	if !c.Completed() {
		t.FailNow()
	}
	var relationships []string
	for _, rel := range c.Graph() {
		relationships = append(relationships, rel.String())
	}
	t.Logf("[Modules] INFO msg=completed graph=[%s]", strings.Join(relationships, ", "))
}

func TestSelfReferentialDependency(t *testing.T) {
	foo := &module{name: "foo", version: "1.0.0"}
	foo.deps = append(foo.deps, foo.Info())
	c := NewContext(ModuleOption(foo), LoggerOption(testLogger{logger: t}))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	c.SetContext(ctx)
	var err error
	go func() {
		err = c.Load(foo.Info())
		cancel()
	}()
	<-c.Done()
	if errors.Is(c.Err(), context.DeadlineExceeded) {
		t.Fatal(c.Err())
	}
	if !errors.Is(err, ErrSelfReferentialDependency) {
		t.Fatal(err)
	}
}

func TestCircularDependency(t *testing.T) {
	baz := &module{name: "baz", version: "1.0.0"}
	baz.deps = append(baz.deps, Info{Name: "foo", Version: "1.0.0"})
	bar := &module{name: "bar", version: "1.0.0"}
	bar.deps = append(bar.deps, baz.Info())
	foo := &module{name: "foo", version: "1.0.0"}
	foo.deps = append(foo.deps, bar.Info())
	c := NewContext(ModuleOption(foo, bar, baz), LoggerOption(testLogger{logger: t}))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	c.SetContext(ctx)
	var err error
	go func() {
		err = c.Load(bar.Info(), foo.Info(), baz.Info())
		cancel()
	}()
	<-c.Done()
	if errors.Is(c.Err(), context.DeadlineExceeded) {
		t.Fatal(c.Err())
	}
	if !errors.Is(err, ErrCircularDependency) {
		t.Fatal(err)
	}
}

func TestInvalid(t *testing.T) {
	foo := &module{name: "foo", version: "1.0.0"}
	c := NewContext(LoggerOption(testLogger{logger: t}))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	c.SetContext(ctx)
	var err error
	go func() {
		err = c.Load(foo.Info())
		cancel()
	}()
	<-c.Done()
	if errors.Is(c.Err(), context.DeadlineExceeded) {
		t.Fatal(c.Err())
	}
	if !errors.Is(err, ErrInvalidModule) {
		t.Fatal(err)
	}
}

type (
	builder struct {
		app *fx.App
	}
)

func (b *builder) Build(svc ...Service) error {
	var opts []fx.Option
	for _, e := range svc {
		opt, ok := e.(fx.Option)
		if !ok {
			return errors.New("invalid type assertion")
		}
		opts = append(opts, opt)
	}
	b.app = fx.New(opts...)
	return nil
}

func TestConfigure(t *testing.T) {
	foo := &module{name: "foo", version: "1.0.0"}
	info := foo.Info()
	foo.services = append(foo.services,
		fx.Decorate(
			fx.Annotate(
				func(bool) bool {
					return true
				},
				fx.OnStart(func(_ context.Context, b bool) error {
					t.Logf("[App] INFO component=%s fx=%t", info, b)
					if !b {
						return errors.New("fx failed")
					}
					return nil
				}),
			),
		),
	)
	c := NewContext(ModuleOption(foo), LoggerOption(testLogger{logger: t}))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	c.SetContext(ctx)
	var err error
	go func() {
		err = c.Load(info)
		cancel()
	}()
	<-c.Done()
	if errors.Is(c.Err(), context.DeadlineExceeded) {
		t.Fatal(c.Err())
	}
	if err != nil {
		t.Fatal(err)
	}
	c.Hook(
		fx.WithLogger(func() fxevent.Logger { return fxtest.NewTestLogger(t) }),
		fx.Provide(
			func() bool {
				return false
			},
		),
		fx.Invoke(func(bool) {}),
	)
	var b builder
	if err = Configure(c, &b); err != nil {
		t.Fatal(err)
	}
	if err = b.app.Start(context.TODO()); err != nil {
		t.Fatal(err)
	}
	if err = b.app.Stop(context.TODO()); err != nil {
		t.Fatal(err)
	}
}

func TestMissingDependency(t *testing.T) {
	foo := &module{name: "foo", version: "1.0.0"}
	foo.deps = append(foo.deps, Info{Name: "bar", Version: "1.0.0"})
	c := NewContext(LoggerOption(testLogger{logger: t}), ModuleOption(foo))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	c.SetContext(ctx)
	var err error
	go func() {
		err = c.Load(foo.Info())
		cancel()
	}()
	<-c.Done()
	if errors.Is(c.Err(), context.DeadlineExceeded) {
		t.Fatal(c.Err())
	}
	if !errors.Is(err, ErrMissingDependency) {
		t.Fatal(err)
	}
}
