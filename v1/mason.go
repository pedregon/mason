package mason

import "errors"

var (
	// ErrInvalidBuilder provides a means for error checking ServiceFunc Builder type casts.
	ErrInvalidBuilder   error = errors.New("invalid builder")
	ErrInvalidContainer error = errors.New("invalid container")
)

type (
	// Builder encourages inversion of control (IoC).
	// https://learn.microsoft.com/en-us/dotnet/architecture/modern-web-apps-azure/architectural-principles#dependency-inversion
	// It is recommended to interface guard.
	// https://caddyserver.com/docs/extending-caddy#interface-guards
	Builder interface {
		// Build constructs a Container.
		Build() (Container, error)
	}
	// Container is a collection of loosely coupled application logic. Empty for future backwards compatability.
	Container any
	// ServiceFunc is a provider function that extends some API. Cast Builder to a struct (optional function).
	// https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
	ServiceFunc func(Builder) error
)

// Configure safely mounts Context ServiceFunc(s) to Builder and returns a built Container.
func Configure[C Container](c *Context, b Builder) (ctn C, err error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, svc := range c.services {
		if err = svc(b); err != nil {
			return
		}
	}
	var e Container
	e, err = b.Build()
	if err != nil {
		return
	}
	var ok bool
	ctn, ok = e.(C)
	if !ok {
		err = ErrInvalidContainer
	}
	return
}
