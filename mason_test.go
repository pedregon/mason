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

package mason_test

import (
	"context"
	"errors"
	"github.com/pedregon/mason/v2"
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
		mu       sync.RWMutex
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
	mort.services = append(mort.services, s...)
	return nil
}

func (mort *nopMortar) list() []mason.Stone {
	mort.mu.RLock()
	defer mort.mu.RUnlock()
	return mort.services
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

func log(t *testing.T, s *mason.Scaffold, ch <-chan mason.Event) {
	for e := range ch {
		if err := e.Err(); err != nil {
			t.Logf("[Mason] ERROR msg='failed to load' info=%s err='%s'", e.Info, err)
		} else {
			t.Logf("[Mason] INFO msg='loaded' module=%s, runtime=%s", e.Info, s.Stat(e.Info))
		}
	}
}

// load conveniently loads all Module(s) asynchronously for Scaffold(ing).
func load(ctx context.Context, cancel context.CancelFunc, s *mason.Scaffold, mod ...mason.Module) (err error) {
	go func() {
		err = s.Load(ctx, mod...)
		cancel()
	}()
	<-ctx.Done()
	if _err := ctx.Err(); errors.Is(_err, context.DeadlineExceeded) {
		return _err
	}
	return
}

func TestModules(t *testing.T) {
	// discover
	baz := &module{name: "baz", version: "1.0.0"}
	bar := &module{name: "bar", version: "1.0.0"}
	bar.deps = append(bar.deps, baz.Info())
	foo := &module{name: "foo", version: "1.0.0"}
	foo.deps = append(foo.deps, bar.Info())
	// register
	modules := []mason.Module{foo, bar, baz}
	// observer
	ch := make(chan mason.Event, 1)
	defer close(ch)
	// construct
	scaffold := mason.New(&nopMortar{}, mason.OnLoad(ch))
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	go log(t, scaffold, ch)
	// hook
	if err := load(ctx, cancel, scaffold, modules...); err != nil {
		t.Fatal(err)
	}
	if mason.Len(scaffold) != len(modules) {
		t.FailNow()
	}
	var relationships []string
	for _, rel := range mason.Graph(scaffold) {
		relationships = append(relationships, rel.String())
	}
	t.Logf("[Modules] INFO msg=completed graph=[%s]", strings.Join(relationships, ", "))
}

func TestSelfReferentialDependency(t *testing.T) {
	// discover
	foo := &module{name: "foo", version: "1.0.0"}
	foo.deps = append(foo.deps, foo.Info())
	// register
	modules := []mason.Module{foo}
	// observer
	ch := make(chan mason.Event, 1)
	defer close(ch)
	// construct
	scaffold := mason.New(&nopMortar{}, mason.OnLoad(ch))
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	go log(t, scaffold, ch)
	// hook
	if err := load(ctx, cancel, scaffold, modules...); !errors.Is(err, mason.ErrSelfReferentialDependency) {
		t.Error(err)
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
	// register
	modules := []mason.Module{foo, bar, baz}
	// observer
	ch := make(chan mason.Event, 1)
	defer close(ch)
	// construct
	scaffold := mason.New(&nopMortar{}, mason.OnLoad(ch))
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	go log(t, scaffold, ch)
	// hook
	if err := load(ctx, cancel, scaffold, modules...); !errors.Is(err, mason.ErrCircularDependency) {
		t.Error(err)
	}
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
	foo.services = append(foo.services, "my service")
	// register
	modules := []mason.Module{foo}
	// observer
	ch := make(chan mason.Event, 1)
	defer close(ch)
	// construct
	mort := &nopMortar{}
	scaffold := mason.New(mort, mason.OnLoad(ch))
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	go log(t, scaffold, ch)
	// hook
	if err := load(ctx, cancel, scaffold, modules...); err != nil {
		t.Error(err)
	}
	services := mort.list()
	if len(services) != len(modules) {
		t.FailNow()
	}
	for _, svc := range services {
		str, ok := svc.(string)
		if !ok {
			t.FailNow()
		}
		t.Log(str)
	}
}

func TestMissingDependency(t *testing.T) {
	// discover
	foo := &module{name: "foo", version: "1.0.0"}
	foo.deps = append(foo.deps, mason.Info{Name: "bar", Version: "1.0.0"})
	// register
	modules := []mason.Module{foo}
	// observer
	ch := make(chan mason.Event, 1)
	defer close(ch)
	// construct
	scaffold := mason.New(&nopMortar{}, mason.OnLoad(ch))
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	go log(t, scaffold, ch)
	// hook
	if err := load(ctx, cancel, scaffold, modules...); !errors.Is(err, mason.ErrMissingDependency) {
		t.Error(err)
	}
}

func skipper(_ mason.Info) bool {
	return true
}

func TestSkipper(t *testing.T) {
	// discover
	foo := &module{name: "foo", version: "1.0.0"}
	// register
	modules := []mason.Module{foo}
	// observer
	ch := make(chan mason.Event, 1)
	defer close(ch)
	// construct
	scaffold := mason.New(&nopMortar{}, mason.OnLoad(ch), mason.SkipOption(skipper))
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	go log(t, scaffold, ch)
	// hook
	if err := load(ctx, cancel, scaffold, modules...); err != nil {
		t.Error(err)
	}
	if scaffold.Stat(foo.Info()).Nanoseconds() > 0 {
		t.FailNow()
	}
}

func TestScaffold_Append(t *testing.T) {
	// discover
	baz := &module{name: "baz", version: "1.0.0"}
	bar := &module{name: "bar", version: "1.0.0"}
	foo := &module{name: "foo", version: "1.0.0"}
	// register
	modulesA := []mason.Module{foo}
	modulesB := []mason.Module{bar, baz, foo}
	// construct
	mort := &nopMortar{}
	scaffoldA := mason.New(mort)
	scaffoldB := mason.New(mort)
	// hook
	ctxA, cancelA := context.WithTimeout(context.TODO(), time.Second)
	if err := load(ctxA, cancelA, scaffoldA, modulesA...); err != nil {
		t.Error(err)
	}
	ctxB, cancelB := context.WithTimeout(context.TODO(), time.Second)
	if err := load(ctxB, cancelB, scaffoldB, modulesB...); err != nil {
		t.Error(err)
	}
	scaffoldA.Append(scaffoldB)
	if mason.Len(scaffoldA) != 3 {
		t.FailNow()
	}
}
