# Mason
![example workflow](https://github.com/pedregon/mason/actions/workflows/main.yml/badge.svg)
Mason is a [compile-time plugin](https://eli.thegreenplace.net/2021/plugins-in-go/) framework for 
[`github.com/uber/fx`](https://uber-go.github.io/fx/). Dependency injection is *messy*. Plugin systems are *messy*. 
This Go module was named after stone masons because it provides a simplistic API for pyramid building applications.
## Design Pattern
The recommended design pattern for implementation is to mimic 
[`database/sql`](https://eli.thegreenplace.net/2019/design-patterns-in-gos-databasesql-package/) with anonymous
package imports. [Build tags](https://www.digitalocean.com/community/tutorials/customizing-go-binaries-with-build-tags) 
may also be used for compile-time inclusivity.
## Contributing
This project is open to [pull requests](https://github.com/pedregon/mason/pulls)!
For discussions, submit an [issue](https://github.com/pedregon/mason/issues). Please 
[sign Git commits](https://docs.github.com/en/authentication/managing-commit-signature-verification/signing-commits) and
adhere to official [Go module versioning](https://go.dev/doc/modules/version-numbers) when 
[publishing](https://go.dev/doc/modules/publishing). Notice the current 
[retractions](https://go.dev/ref/mod#go-mod-file-retract) in the [go.mod](https://proxy.golang.org/).
### v2
Why? Mason was developed to offer an alternative to the standard library, [`plugin`](https://pkg.go.dev/plugin),
RPC solutions such as [`github.com/hashicorp/go-plugin`](https://github.com/hashicorp/go-plugin),
and other network-based solutions. Possibly the most contentious aspect of
the API is the use of [`github.com/uber/fx`](https://uber-go.github.io/fx/),
a dependency injection system which to some is considered an unecessary anti-pattern. 
On the other hand, it is the best documented and easiest to use of its class. The next version
of Mason may go in two different directions. New versions could involve
supporting other application building libraries or even evolve into a mature abstraction.
Mason by no means claims to be a perfect solution, and it is open to feedback!