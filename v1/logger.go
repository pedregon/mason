package mason

type (
	Logger interface {
		Info(msg string, info Info, kv ...KV)
		Error(msg string, info Info, err error)
	}
	KV struct {
		Key   string
		Value string
	}
	nopLogger struct{}
)

func (nopLogger) Info(msg string, info Info, kv ...KV) {}

func (nopLogger) Error(msg string, info Info, err error) {}
