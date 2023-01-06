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
	"runtime/debug"
	"strings"
	"sync"
	"testing"
	"time"
)

var (
	_ Registrar = (*nopRegistrar)(nil)
	_ Module    = (*module)(nil)
)

type (
	nopRegistrar struct {
		mu       sync.Mutex
		services []Service
	}
	module struct {
		name     string
		version  string
		deps     []Info
		services []Service
	}
)

func (reg *nopRegistrar) Register(svc Service) error {
	reg.mu.Lock()
	defer reg.mu.Unlock()
	reg.services = append(reg.services, svc)
	return nil
}

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
		if err = c.Hook(mod.services...); err != nil {
			return
		}
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
	reg := new(nopRegistrar)
	baz := &module{name: "baz", version: "1.0.0"}
	bar := &module{name: "bar", version: "1.0.0"}
	bar.deps = append(bar.deps, baz.Info())
	foo := &module{name: "foo", version: "1.0.0"}
	foo.deps = append(foo.deps, bar.Info())
	c := NewContext(reg, ModuleOption(foo, bar, baz), LoggerOption(testLogger{logger: t}))
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
	reg := new(nopRegistrar)
	foo := &module{name: "foo", version: "1.0.0"}
	foo.deps = append(foo.deps, foo.Info())
	c := NewContext(reg, ModuleOption(foo), LoggerOption(testLogger{logger: t}))
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
	reg := new(nopRegistrar)
	baz := &module{name: "baz", version: "1.0.0"}
	baz.deps = append(baz.deps, Info{Name: "foo", Version: "1.0.0"})
	bar := &module{name: "bar", version: "1.0.0"}
	bar.deps = append(bar.deps, baz.Info())
	foo := &module{name: "foo", version: "1.0.0"}
	foo.deps = append(foo.deps, bar.Info())
	c := NewContext(reg, ModuleOption(foo, bar, baz), LoggerOption(testLogger{logger: t}))
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
	reg := new(nopRegistrar)
	foo := &module{name: "foo", version: "1.0.0"}
	c := NewContext(reg, LoggerOption(testLogger{logger: t}))
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

var (
	_ Registrar = (*fxRegistrar)(nil)
)

type (
	fxRegistrar struct {
		mu      sync.RWMutex
		options []fx.Option
	}
)

func (reg *fxRegistrar) Register(svc Service) error {
	reg.mu.Lock()
	defer reg.mu.Unlock()
	reg.options = append(reg.options, svc.(fx.Option))
	return nil
}

func (reg *fxRegistrar) Construct() *fx.App {
	reg.mu.RLock()
	defer reg.mu.RUnlock()
	return fx.New(reg.options...)
}

func TestRegistrar(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Logf("panic occurred: %s", err)
			t.Logf("stack trace: %s", debug.Stack())
			t.FailNow()
		}
	}()
	reg := new(fxRegistrar)
	foo := &module{name: "foo", version: "1.0.0"}
	info := foo.Info()
	var err error
	if err = reg.Register(
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
	); err != nil {
		t.Fatal(err)
	}
	c := NewContext(reg, ModuleOption(foo), LoggerOption(testLogger{logger: t}))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	c.SetContext(ctx)
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
	if err = c.Hook(
		fx.WithLogger(func() fxevent.Logger { return fxtest.NewTestLogger(t) }),
		fx.Provide(
			func() bool {
				return false
			},
		),
		fx.Invoke(func(bool) {}),
	); err != nil {
		t.Fatal(err)
	}
	app := reg.Construct()
	if err = app.Start(context.TODO()); err != nil {
		t.Fatal(err)
	}
	if err = app.Stop(context.TODO()); err != nil {
		t.Fatal(err)
	}
}

func TestMissingDependency(t *testing.T) {
	reg := new(nopRegistrar)
	foo := &module{name: "foo", version: "1.0.0"}
	foo.deps = append(foo.deps, Info{Name: "bar", Version: "1.0.0"})
	c := NewContext(reg, LoggerOption(testLogger{logger: t}), ModuleOption(foo))
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

func TestNilLogger(t *testing.T) {
	reg := new(nopRegistrar)
	baz := &module{name: "baz", version: "1.0.0"}
	bar := &module{name: "bar", version: "1.0.0"}
	bar.deps = append(bar.deps, baz.Info())
	foo := &module{name: "foo", version: "1.0.0"}
	foo.deps = append(foo.deps, bar.Info())
	c := NewContext(reg, ModuleOption(foo, bar, baz))
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
