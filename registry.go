package roster

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// Factory is a function that creates a new instance of T from a URI.
type Factory[T any] func(ctx context.Context, uri string) (T, error)

// Registry is a type-safe lookup table mapping URI schemes to `Factory[T]` functions.
type Registry[T any] struct {
	roster Roster
}

// NewRegistry returns a new `Registry` instance for creating values of type T.
func NewRegistry[T any]() (*Registry[T], error) {

	r, err := NewDefaultRoster()

	if err != nil {
		return nil, err
	}

	return &Registry[T]{roster: r}, nil
}

// Register associates 'scheme' with 'f'.
func (r *Registry[T]) Register(ctx context.Context, scheme string, f Factory[T]) error {

	if f == nil {
		return fmt.Errorf("Nothing to register for scheme '%s'", scheme)
	}

	return r.roster.Register(ctx, scheme, f)
}

// MustRegister is like `Register` but panics if registration fails. It is intended for use in `init` functions.
func (r *Registry[T]) MustRegister(ctx context.Context, scheme string, f Factory[T]) {

	if err := r.Register(ctx, scheme, f); err != nil {
		panic(fmt.Sprintf("Failed to register '%s': %v", scheme, err))
	}
}

// New parses 'uri' and invokes the `Factory` registered for its scheme.
func (r *Registry[T]) New(ctx context.Context, uri string) (T, error) {

	var zero T

	u, err := url.Parse(uri)

	if err != nil {
		return zero, fmt.Errorf("Failed to parse URI, %w", err)
	}

	d, err := r.roster.Driver(ctx, u.Scheme)

	if err != nil {
		return zero, fmt.Errorf("Failed to find registration for %s, %w", u.Scheme, err)
	}

	f, ok := d.(Factory[T])

	if !ok {
		return zero, fmt.Errorf("Invalid registration type for %s: %T", u.Scheme, d)
	}

	return f(ctx, uri)
}

// Schemes returns the sorted list of registered schemes, each formatted as "scheme://".
func (r *Registry[T]) Schemes(ctx context.Context) []string {

	drivers := r.roster.Drivers(ctx)
	schemes := make([]string, len(drivers))

	for i, d := range drivers {
		schemes[i] = strings.ToLower(d) + "://"
	}

	return schemes
}
