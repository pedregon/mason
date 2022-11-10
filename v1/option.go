package mason

type (
	Option func(*Context)
)

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

func LoggerOption(logger Logger) Option {
	return func(c *Context) {
		c.logger = logger
	}
}
