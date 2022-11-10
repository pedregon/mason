module github.com/pedregon/mason

go 1.19

retract (
	v1.0.0 // Published accidentally
	v1.0.1 // Contains retractions only
)

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/dig v1.15.0 // indirect
	go.uber.org/fx v1.18.2 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	go.uber.org/zap v1.23.0 // indirect
	golang.org/x/sys v0.2.0 // indirect
)
