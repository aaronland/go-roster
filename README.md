# go-roster

Go package provides interfaces and methods for defining internal lookup tables (or "rosters") for registering and instantiatinge custom interfaces with multiple implementations.

## Documentation

[![Go Reference](https://pkg.go.dev/badge/github.com/aaronland/go-roster/v2.svg)](https://pkg.go.dev/github.com/aaronland/go-roster/v2)

## Example (v2, generics)

Define an interface and a registry for it. This is the only boilerplate a consuming package needs:

```
package greeter

import (
	"context"

	"github.com/aaronland/go-roster/v2"
)

type Greeter interface {
	Greet() string
}

var registry, _ = roster.NewRegistry[Greeter]()

// GreeterInitializationFunc is an alias, so implementations can use either name.
type GreeterInitializationFunc = roster.Factory[Greeter]

func RegisterGreeter(ctx context.Context, scheme string, f GreeterInitializationFunc) error {
	return registry.Register(ctx, scheme, f)
}

func NewGreeter(ctx context.Context, uri string) (Greeter, error) {
	return registry.New(ctx, uri)
}

func Schemes() []string {
	return registry.Schemes(context.Background())
}
```

Implement the interface and register it (use `MustRegister` to panic on failure, for example in `init`):

```
package greeter

import (
	"context"
	"net/url"
)

type PathGreeter string

func (g PathGreeter) Greet() string {
	return "Hello, " + string(g)
}

func init() {
	registry.MustRegister(context.Background(), "path", NewPathGreeter)
}

func NewPathGreeter(ctx context.Context, uri string) (Greeter, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	return PathGreeter(u.Path), nil
}
```

Create instances from URIs:

```
g, err := greeter.NewGreeter(ctx, "path:///world")
fmt.Println(g.Greet())        // Hello, /world

fmt.Println(greeter.Schemes()) // [path://]
```

`Registry[T]` handles URI parsing, scheme lookup, type checking and scheme listing. A working version is in [registry_test.go](registry_test.go).

The lower-level `Roster` interface (see [roster_test.go](roster_test.go)) is still available.

## Migrating from v1

* Import `github.com/aaronland/go-roster/v2`. Requires Go 1.18+.
* `Registry[T]` is new; the `Roster` and `DefaultRoster` APIs are unchanged.

## Concrete examples

* https://github.com/whosonfirst/go-reader
* https://github.com/whosonfirst/go-writer