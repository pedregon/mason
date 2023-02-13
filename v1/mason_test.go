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

package mason_test

import (
	"context"
	"errors"
	"github.com/pedregon/mason/v1"
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
	_ mason.Mortar = (*nopMortar)(nil)
	_ mason.Module = (*module)(nil)
)

type (
	nopMortar struct {
		mu       sync.Mutex
		services []mason.Stone
	}
	module struct {
		name     string
		version  string
		deps     []mason.Info
		services []mason.Stone
	}
)

func (mort *nopMortar) Hook(s ...mason.Stone) error {
	mort.mu.Lock()
	defer mort.mu.Unlock()
	mort.services = append(mort.services, s)
	return nil
}

func (mod module) Info() (info mason.Info) {
	info.Name = mod.name
	info.Version = mod.version
	return
}

func (mod module) Provision(c *mason.Context) (err error) {
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

func logger(t *testing.T) mason.Observer {
	return func(c *mason.Context, ch <-chan mason.Event) {
		go func() {
			for _ = range ch {
				//info := e.Blame()
				//if err := e.Err(); err != nil {
				//	t.Logf("[Modules] ERROR msg='failed to load' info=%s err='%s'", info, err.Error())
				//} else {
				//	t.Logf("[Modules] INFO msg='loaded' module=%s, runtime=%s", info, c.Stat(info))
				//}
			}
		}()
	}
}

func TestModules(t *testing.T) {
	// discover
	baz := &module{name: "baz", version: "1.0.0"}
	bar := &module{name: "bar", version: "1.0.0"}
	bar.deps = append(bar.deps, baz.Info())
	foo := &module{name: "foo", version: "1.0.0"}
	foo.deps = append(foo.deps, bar.Info())
	c, cancel := mason.NewContext(context.TODO(), &nopMortar{}, mason.TimeoutOption(time.Second),
		mason.WatchOption(logger(t)))
	var err error
	// register
	if err = c.Register(foo, bar, baz); err != nil {
		t.Fatal(err)
	}
	// hook
	go func() {
		err = mason.Load(c)
		cancel()
	}()
	<-c.Done()
	if errors.Is(c.Err(), context.DeadlineExceeded) {
		t.Fatal(c.Err())
	}
	if err != nil {
		t.Fatal(err)
	}
	if len(mason.Loaded(c)) != mason.Len(c) {
		t.FailNow()
	}
	var relationships []string
	for _, rel := range mason.Graph(c) {
		relationships = append(relationships, rel.String())
	}
	t.Logf("[Modules] INFO msg=completed graph=[%s]", strings.Join(relationships, ", "))
}

func TestSelfReferentialDependency(t *testing.T) {
	// discover
	foo := &module{name: "foo", version: "1.0.0"}
	foo.deps = append(foo.deps, foo.Info())
	c, cancel := mason.NewContext(context.TODO(), &nopMortar{}, mason.TimeoutOption(time.Second),
		mason.WatchOption(logger(t)))
	var err error
	// register
	if err = c.Register(foo); err != nil {
		t.Fatal(err)
	}
	// hook
	go func() {
		err = mason.Load(c)
		cancel()
	}()
	<-c.Done()
	if errors.Is(c.Err(), context.DeadlineExceeded) {
		t.Fatal(c.Err())
	}
	if !errors.Is(err, mason.ErrSelfReferentialDependency) {
		t.Fatal(err)
	}
}

func TestCircularDependency(t *testing.T) {
	// discover
	baz := &module{name: "baz", version: "1.0.0"}
	baz.deps = append(baz.deps, mason.Info{Name: "foo", Version: "1.0.0"})
	bar := &module{name: "bar", version: "1.0.0"}
	bar.deps = append(bar.deps, baz.Info())
	foo := &module{name: "foo", version: "1.0.0"}
	foo.deps = append(foo.deps, bar.Info())
	c, cancel := mason.NewContext(context.TODO(), &nopMortar{}, mason.TimeoutOption(time.Second),
		mason.WatchOption(logger(t)))
	var err error
	// register
	if err = c.Register(foo, bar, baz); err != nil {
		t.Fatal(err)
	}
	// hook
	go func() {
		err = mason.Load(c)
		cancel()
	}()
	<-c.Done()
	if errors.Is(c.Err(), context.DeadlineExceeded) {
		t.Fatal(c.Err())
	}
	if !errors.Is(err, mason.ErrCircularDependency) {
		t.Fatal(err)
	}
}

func TestInvalid(t *testing.T) {
	// discover
	foo := &module{name: "foo", version: "1.0.0"}
	c, cancel := mason.NewContext(context.TODO(), &nopMortar{}, mason.TimeoutOption(time.Second),
		mason.WatchOption(logger(t)))
	var err error
	// hook
	go func() {
		err = c.Load(foo.Info())
		cancel()
	}()
	<-c.Done()
	if errors.Is(c.Err(), context.DeadlineExceeded) {
		t.Fatal(c.Err())
	}
	if !errors.Is(err, mason.ErrInvalidModule) {
		t.Fatal(err)
	}
}

var (
	_ mason.Mortar = (*fxMortar)(nil)
)

type (
	fxMortar struct {
		mu      sync.RWMutex
		options []fx.Option
	}
)

func (mort *fxMortar) Hook(s ...mason.Stone) error {
	mort.mu.Lock()
	defer mort.mu.Unlock()
	for _, e := range s {
		mort.options = append(mort.options, e.(fx.Option))
	}
	return nil
}

func (mort *fxMortar) Trowel() *fx.App {
	mort.mu.RLock()
	defer mort.mu.RUnlock()
	return fx.New(mort.options...)
}

func TestMortar(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Logf("panic occurred: %s", err)
			t.Logf("stack trace: %s", debug.Stack())
			t.FailNow()
		}
	}()
	// discover
	foo := &module{name: "foo", version: "1.0.0"}
	mort := &fxMortar{}
	var err error
	if err = mort.Hook(
		fx.Decorate(
			fx.Annotate(
				func(bool) bool {
					return true
				},
				fx.OnStart(func(_ context.Context, b bool) error {
					t.Logf("[App] INFO component=%s fx=%t", foo.Info(), b)
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
	c, cancel := mason.NewContext(context.TODO(), mort, mason.TimeoutOption(time.Second),
		mason.WatchOption(logger(t)))
	// register
	if err = c.Register(foo); err != nil {
		t.Fatal(err)
	}
	// hook
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
	go func() {
		err = mason.Load(c)
		cancel()
		cancel()
	}()
	<-c.Done()
	if errors.Is(c.Err(), context.DeadlineExceeded) {
		t.Fatal(c.Err())
	}
	if err != nil {
		t.Fatal(err)
	}
	app := mort.Trowel()
	if err = app.Start(context.TODO()); err != nil {
		t.Fatal(err)
	}
	if err = app.Stop(context.TODO()); err != nil {
		t.Fatal(err)
	}
}

func TestMissingDependency(t *testing.T) {
	// discover
	foo := &module{name: "foo", version: "1.0.0"}
	foo.deps = append(foo.deps, mason.Info{Name: "bar", Version: "1.0.0"})
	c, cancel := mason.NewContext(context.TODO(), &nopMortar{}, mason.TimeoutOption(time.Second),
		mason.WatchOption(logger(t)))
	var err error
	// register
	if err = c.Register(foo); err != nil {
		t.Fatal(err)
	}
	// hook
	go func() {
		err = c.Load(foo.Info())
		cancel()
	}()
	<-c.Done()
	if errors.Is(c.Err(), context.DeadlineExceeded) {
		t.Fatal(c.Err())
	}
	if !errors.Is(err, mason.ErrMissingDependency) {
		t.Fatal(err)
	}
}

func TestDuplicateModule(t *testing.T) {
	// discover
	foo := &module{name: "foo", version: "1.0.0"}
	c, cancel := mason.NewContext(context.TODO(), &nopMortar{}, mason.WatchOption(logger(t)))
	defer cancel()
	// register
	err := c.Register(foo, foo)
	if !errors.Is(err, mason.ErrDuplicateModule) {
		t.FailNow()
	}
}

func skipper(_ mason.Info) bool {
	return true
}

func TestSkipper(t *testing.T) {
	// discover
	foo := &module{name: "foo", version: "1.0.0"}
	c, cancel := mason.NewContext(context.TODO(), &nopMortar{}, mason.TimeoutOption(time.Second),
		mason.SkipOption(skipper))
	var err error
	// register
	if err = c.Register(foo); err != nil {
		t.Fatal(err)
	}
	// hook
	go func() {
		err = mason.Load(c)
		cancel()
	}()
	<-c.Done()
	if errors.Is(c.Err(), context.DeadlineExceeded) {
		t.Fatal(c.Err())
	}
	if err != nil {
		t.Fatal(err)
	}
	if len(mason.Loaded(c)) > 0 {
		t.FailNow()
	}
}
