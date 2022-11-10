# Mason
Mason is a [compile-time plugin](https://eli.thegreenplace.net/2021/plugins-in-go/) framework for 
[`github.com/uber/fx`](https://uber-go.github.io/fx/). Dependency injection is *messy*. Plugin systems are *messy*. 
This module was named after stone masons because it provides a simplistic API for pyramid building applications.
## Design Pattern
The recommended design pattern for implementation is to mimic 
[`database/sql`](https://eli.thegreenplace.net/2019/design-patterns-in-gos-databasesql-package/) with anonymous
package imports. [Build tags](https://www.digitalocean.com/community/tutorials/customizing-go-binaries-with-build-tags) 
may also be used for compile-time inclusivity.