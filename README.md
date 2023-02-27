# Mason
[![test](https://github.com/pedregon/mason/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/pedregon/mason/actions/workflows/test.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/pedregon/mason.svg)](https://pkg.go.dev/github.com/pedregon/mason)

<img src="https://upload.wikimedia.org/wikipedia/commons/thumb/a/af/All_Gizah_Pyramids.jpg/580px-All_Gizah_Pyramids.jpg" width="350" height="250" />

Mason is a [plugin](https://eli.thegreenplace.net/2021/plugins-in-go/) framework for statically compiled Go.
Plugin systems are *messy*.
[Inversion of control](https://www.henrydu.com/2022/01/09/golang-inversion-of-control/) is *messy*.
This Go module was named after stone masons because it aims to provide a simplistic API for constructing
application pyramids.
## Installation
```
go get -u github.com/pedregon/mason
```
## Examples
There may be better examples in a future revision, but for now check out
[TestModules](https://github.com/pedregon/mason/raw/main/v2/mason_test.go) or
[TestMortar](https://github.com/pedregon/mason/raw/main/v2/mason_test.go).
## Design Pattern
The recommended design pattern for plugin registration is to mimic
[`database/sql`](https://eli.thegreenplace.net/2019/design-patterns-in-gos-databasesql-package/) with anonymous
package imports as the plugin discovery mechanism.
[Build tags](https://www.digitalocean.com/community/tutorials/customizing-go-binaries-with-build-tags)
may also be used for compile-time inclusivity.
## Rational
Mason was developed to offer an alternative to the Go standard library, [`plugin`](https://pkg.go.dev/plugin),
RPC solutions such as [`github.com/hashicorp/go-plugin`](https://github.com/hashicorp/go-plugin),
and network-based solutions. Mason works by constructing
[`Scaffold(ing)`](https://github.com/pedregon/mason/blob/main/v2/scaffold.go) to apply
[`Mortar`](https://github.com/pedregon/mason/blob/main/v2/mason.go) on 
[`Stone`](https://github.com/pedregon/mason/blob/main/v2/mason.go) from 
[`Module(s)`](https://github.com/pedregon/mason/blob/main/v1/module.go). On `Module` load, `Scaffold` creates a 
[`Context`](https://github.com/pedregon/mason/blob/main/v1/context.go) for hooking API logic. 
`Module(s)` are hooked by the `Context` and
provide `Stone` as building blocks that extend the API. Mason takes care of the plugin dependency plumbing and 
empowers custom discovery, registration, and hooking. `Mortar` is the underlying glue.
Initially, [`github.com/uber/fx`](https://uber-go.github.io/fx/) was first-class
supported as the inversion of control mechanism because it was arguably the best documented and easiest to use
(no code generation or type inferring) dependency injection framework, but it has since been decided that
a) users do not want to be forcibly dependent on an external library and
b) some consider dependency injection systems to be an unnecessary anti-pattern. Therefore, to remain idiomatic,
`Mortar` inversion of control via `Stone` hooking was abstracted. `Context`, itself, even implements the `Mortar` 
interface to foster chain of responsibility. Mason by no means claims to be a perfect solution, 
and is open to feedback!
## Contributing
This project is open to [pull requests](https://github.com/pedregon/mason/pulls)!
For discussions, submit an [issue](https://github.com/pedregon/mason/issues). Please
[sign Git commits](https://docs.github.com/en/authentication/managing-commit-signature-verification/signing-commits) and
adhere to official [Go module versioning](https://go.dev/doc/modules/version-numbers) when
[publishing](https://go.dev/doc/modules/publishing). Notice the current
[retractions](https://go.dev/ref/mod#go-mod-file-retract) in the [go.mod](https://proxy.golang.org/).
