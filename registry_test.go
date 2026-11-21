package roster

import (
	"context"
	"net/url"
	"reflect"
	"testing"
)

type Greeter interface {
	Greet() string
}

type pathGreeter string

func (g pathGreeter) Greet() string { return string(g) }

func newPathGreeter(ctx context.Context, uri string) (Greeter, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	return pathGreeter(u.Path), nil
}

func TestRegistry(t *testing.T) {

	ctx := context.Background()

	r, err := NewRegistry[Greeter]()

	if err != nil {
		t.Fatalf("Failed to create registry, %v", err)
	}

	err = r.Register(ctx, "Path", newPathGreeter)

	if err != nil {
		t.Fatalf("Failed to register, %v", err)
	}

	if r.Register(ctx, "path", newPathGreeter) == nil {
		t.Fatal("Expected duplicate registration to fail")
	}

	g, err := r.New(ctx, "path:///hello")

	if err != nil {
		t.Fatalf("Failed to create greeter, %v", err)
	}

	if g.Greet() != "/hello" {
		t.Fatalf("Unexpected result: '%s'", g.Greet())
	}

	_, err = r.New(ctx, "nope://")

	if err == nil {
		t.Fatal("Expected unknown scheme to fail")
	}

	expected := []string{"path://"}

	if !reflect.DeepEqual(r.Schemes(ctx), expected) {
		t.Fatalf("Unexpected schemes: %v", r.Schemes(ctx))
	}
}

func TestRegistryMustRegister(t *testing.T) {

	ctx := context.Background()

	r, _ := NewRegistry[Greeter]()

	r.MustRegister(ctx, "path", newPathGreeter)

	defer func() {
		if recover() == nil {
			t.Fatal("Expected MustRegister to panic on duplicate registration")
		}
	}()

	r.MustRegister(ctx, "path", newPathGreeter)
}

func TestRegistryWrongType(t *testing.T) {

	ctx := context.Background()

	r, _ := NewRegistry[Greeter]()

	// Bypass the typed API to confirm a bad registration returns an error rather than panicking.
	_ = r.roster.Register(ctx, "bad", "not a factory")

	if _, err := r.New(ctx, "bad://"); err == nil {
		t.Fatal("Expected error for invalid registration type")
	}
}
