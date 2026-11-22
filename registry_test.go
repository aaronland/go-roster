package roster

import (
	"context"
	"io"
	"os"
	"reflect"
	"sync"
	"testing"
)

// Writer is the example interface that implementations are registered for.
type Writer interface {
	Write(ctx context.Context, data []byte) error
}

// NullWriter discards everything written to it.
type NullWriter struct{}

func NewNullWriter(ctx context.Context, uri string) (Writer, error) {
	return &NullWriter{}, nil
}

func (w *NullWriter) Write(ctx context.Context, data []byte) error {
	_, err := io.Discard.Write(data)
	return err
}

// StdoutWriter writes to os.Stdout.
type StdoutWriter struct{}

func NewStdoutWriter(ctx context.Context, uri string) (Writer, error) {
	return &StdoutWriter{}, nil
}

func (w *StdoutWriter) Write(ctx context.Context, data []byte) error {
	_, err := os.Stdout.Write(data)
	return err
}

func newTestRegistry(t *testing.T) *Registry[Writer] {

	t.Helper()

	ctx := context.Background()
	r := NewRegistry[Writer]()

	r.MustRegister(ctx, "null", NewNullWriter)
	r.MustRegister(ctx, "stdout", NewStdoutWriter)

	return r
}

func TestRegistry(t *testing.T) {

	ctx := context.Background()
	r := newTestRegistry(t)

	w, err := r.New(ctx, "null://")

	if err != nil {
		t.Fatalf("Failed to create null writer, %v", err)
	}

	if _, ok := w.(*NullWriter); !ok {
		t.Fatalf("Expected *NullWriter, got %T", w)
	}

	if err := w.Write(ctx, []byte("hello")); err != nil {
		t.Fatalf("Failed to write, %v", err)
	}

	// Schemes are case-insensitive
	w, err = r.New(ctx, "stdout://")

	if err != nil {
		t.Fatalf("Failed to create stdout writer, %v", err)
	}

	if _, ok := w.(*StdoutWriter); !ok {
		t.Fatalf("Expected *StdoutWriter, got %T", w)
	}

	if _, err = r.New(ctx, "nope://"); err == nil {
		t.Fatal("Expected unknown scheme to fail")
	}

	if _, err = r.New(ctx, "://bad"); err == nil {
		t.Fatal("Expected unparseable URI to fail")
	}

	expected := []string{"null://", "stdout://"}

	if !reflect.DeepEqual(r.Schemes(ctx), expected) {
		t.Fatalf("Unexpected schemes: %v", r.Schemes(ctx))
	}
}

func TestStdoutWriter(t *testing.T) {

	ctx := context.Background()
	r := newTestRegistry(t)

	w, err := r.New(ctx, "stdout://")

	if err != nil {
		t.Fatalf("Failed to create stdout writer, %v", err)
	}

	// Redirect os.Stdout to a pipe to capture what the writer emits.
	pr, pw, err := os.Pipe()

	if err != nil {
		t.Fatalf("Failed to create pipe, %v", err)
	}

	orig := os.Stdout
	os.Stdout = pw

	defer func() { os.Stdout = orig }()

	err = w.Write(ctx, []byte("hello"))

	os.Stdout = orig
	pw.Close()

	if err != nil {
		t.Fatalf("Failed to write, %v", err)
	}

	body, err := io.ReadAll(pr)

	if err != nil {
		t.Fatalf("Failed to read pipe, %v", err)
	}

	if string(body) != "hello" {
		t.Fatalf("Unexpected output: '%s'", string(body))
	}
}

func TestRegistryRegister(t *testing.T) {

	ctx := context.Background()
	r := newTestRegistry(t)

	if r.Register(ctx, "NULL", NewNullWriter) == nil {
		t.Fatal("Expected duplicate registration to fail")
	}

	if r.Register(ctx, "nil", nil) == nil {
		t.Fatal("Expected nil factory to fail")
	}
}

func TestRegistryMustRegister(t *testing.T) {

	ctx := context.Background()
	r := newTestRegistry(t)

	defer func() {
		if recover() == nil {
			t.Fatal("Expected MustRegister to panic on duplicate registration")
		}
	}()

	r.MustRegister(ctx, "null", NewNullWriter)
}

func TestRegistryConcurrent(t *testing.T) {

	ctx := context.Background()
	r := newTestRegistry(t)

	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = r.New(ctx, "null://")
			_ = r.Schemes(ctx)
		}()
	}

	wg.Wait()
}
