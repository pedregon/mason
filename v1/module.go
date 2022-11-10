package mason

import (
	"errors"
	"github.com/cespare/xxhash/v2"
)

var (
	ErrInvalidModule             error = errors.New("invalid module")
	ErrSelfReferentialDependency error = errors.New("self-referential module dependency")
	ErrCircularDependency        error = errors.New("circular module dependency")
)

type (
	// Info is Module information.
	Info struct {
		Name    string
		Version string
	}
	// Module is a compile-time plugin based on https://caddyserver.com/docs/extending-caddy.
	Module interface {
		// Info returns identifying information.
		Info() Info
		// Provision initializes.
		Provision(c *Context) error
	}
	// Dependency is a Module dependency relationship.
	Dependency struct {
		From Info
		To   Info
	}
)

// ID implements graph.Node from https://github.com/gonum/gonum.
func (i Info) ID() int64 {
	return int64(xxhash.Sum64String(i.String()))
}

// String implements fmt.Stringer.
func (i Info) String() string {
	return i.Name + "-" + i.Version
}

// String implements fmt.Stringer.
func (d Dependency) String() string {
	return d.To.String() + " <= " + d.From.String()
}
