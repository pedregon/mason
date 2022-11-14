# Mason
Mason is a [compile-time plugin](https://eli.thegreenplace.net/2021/plugins-in-go/) framework for 
[`github.com/uber/fx`](https://uber-go.github.io/fx/). Dependency injection is *messy*. Plugin systems are *messy*. 
This module was named after stone masons because it provides a simplistic API for pyramid building applications.
## Design Pattern
The recommended design pattern for implementation is to mimic 
[`database/sql`](https://eli.thegreenplace.net/2019/design-patterns-in-gos-databasesql-package/) with anonymous
package imports. [Build tags](https://www.digitalocean.com/community/tutorials/customizing-go-binaries-with-build-tags) 
may also be used for compile-time inclusivity.
## v2
Why? Mason was developed to offer an alternative to the standard library, `plugin`, RPC solution such as, 
`github.com/hashicorp/go-plugin`, and various network-based solutions. Possibly the most contentious aspect of
this Go package is the use of `github.com/uber/fx`, a dependency injection system which to some is considered
an unecessary anti-pattern. , 
## Contributing
This project is open to pull requests!
- Sign Git Commits
- Submit pull request