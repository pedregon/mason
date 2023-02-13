module github.com/pedregon/mason

go 1.19

retract (
	v1.0.1 // Contains retractions only
	v1.0.0 // Published accidentally
)

require go.uber.org/fx v1.19.0

require (
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/dig v1.16.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.23.0 // indirect
	golang.org/x/sys v0.0.0-20210903071746-97244b99971b // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
