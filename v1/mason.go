package mason

type (
	// Builder achieves inversion of control (IoC).
	Builder interface {
		// Build mounts Service(s) to some API.
		Build(...Service) error
	}
	// Service is a provider that extends some API.
	Service any
)

// Configure safely registers Context Service(s) with a Builder.
func Configure(c *Context, b Builder) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return b.Build(c.services)
}
