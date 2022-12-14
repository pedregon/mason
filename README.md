# Mason
[![test](https://github.com/pedregon/mason/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/pedregon/mason/actions/workflows/test.yml)

![pyramids](https://www.lol-smurfs.com/blog/wp-content/uploads/2017/01/21.jpg)

Mason is a [plugin](https://eli.thegreenplace.net/2021/plugins-in-go/) framework for statically compiled Go.
Plugin systems are *messy*.
[Inversion of control](https://www.henrydu.com/2022/01/09/golang-inversion-of-control/) is *messy*.
This Go module was named after stone masons because it provides a simplistic API for constructing application pyramids.
There may be better examples in a future update, but for now check out
[TestMortar](https://github.com/pedregon/mason/blob/main/v1/mason_test.go) for an
[`github.com/uber/fx`](https://uber-go.github.io/fx/) implementation.
## Installation
```
go get -u github.com/pedregon/mason
```
## Design Pattern
The recommended design pattern for plugin registration is to mimic
[`database/sql`](https://eli.thegreenplace.net/2019/design-patterns-in-gos-databasesql-package/) with anonymous
package imports as the plugin discovery mechanism.
[Build tags](https://www.digitalocean.com/community/tutorials/customizing-go-binaries-with-build-tags)
may also be used for compile-time inclusivity.
## Rational
Mason was developed to offer an alternative to the standard library [`plugin`](https://pkg.go.dev/plugin),
RPC solutions such as [`github.com/hashicorp/go-plugin`](https://github.com/hashicorp/go-plugin),
and other network-based solutions. Mason revolves around a
[`Context`](https://github.com/pedregon/mason/blob/main/v1/context.go) using
[`Mortar`](https://github.com/pedregon/mason/blob/main/v1/mason.go) that glues
[`Module`](https://github.com/pedregon/mason/blob/main/v1/module.go) extending logic to application extensible logic. 
*Loaded* `Module(s)` provide [`Stone`](https://github.com/pedregon/mason/blob/main/v1/mason.go) building blocks that are
hooked by the `Context`. Mason takes care of the plugin dependency plumbing and empowers custom discovery,
registration, and hooking. Initially, [`github.com/uber/fx`](https://uber-go.github.io/fx/) was first-class
supported as the inversion of control mechanism because it was arguably the best documented and easiest to use
(no code generation or type inferring) dependency injection framework, but it has since been decided that
a) users do not want to be forcibly dependent on an external library and
b) some consider dependency injection systems to be an unnecessary anti-pattern. Therefore, to remain idiomatic,
`Mortar` inversion of control via `Stone` hooking was abstracted. `Context`, itself, even implements the `Mortar` 
interface to foster some wondrous chain of responsibility. Mason by no means claims to be a perfect solution, 
and is open to feedback!
## Contributing
This project is open to [pull requests](https://github.com/pedregon/mason/pulls)!
For discussions, submit an [issue](https://github.com/pedregon/mason/issues). Please
[sign Git commits](https://docs.github.com/en/authentication/managing-commit-signature-verification/signing-commits) and
adhere to official [Go module versioning](https://go.dev/doc/modules/version-numbers) when
[publishing](https://go.dev/doc/modules/publishing). Notice the current
[retractions](https://go.dev/ref/mod#go-mod-file-retract) in the [go.mod](https://proxy.golang.org/).
