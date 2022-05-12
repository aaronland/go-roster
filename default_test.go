package roster

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Something that implements the Example interface

type TimeExample struct {
	Example
	time time.Time
}

func NewTimeExample(ctx context.Context, uri string) (Example, error) {

	now := time.Now()

	s := &TimeExample{
		time: now,
	}

	return s, nil
}

func (e *TimeExample) String() string {
	return e.time.Format(time.RFC3339)
}

func TestDefaultRoster(t *testing.T) {

	ctx := context.Background()

	dr, err := NewDefaultRoster()

	if err != nil {
		t.Fatalf("Failed to create new default roster, %v", err)
	}

	// START OF a little bit of indirection
	// to ensure the Go compiler has the right interface definitions

	register := func(ctx context.Context, scheme string, init_func ExampleInitializationFunc) error {
		return dr.Register(ctx, scheme, init_func)
	}

	err = register(ctx, "time", NewTimeExample)

	// END OF a little bit of indirection

	if err != nil {
		t.Fatalf("Failed to register example, %v", err)
	}

	i, err := dr.Driver(ctx, "time")

	if err != nil {
		t.Fatalf("Failed to retrieve entry for 'time', %v", err)
	}

	init_func := i.(ExampleInitializationFunc)

	t_ex, err := init_func(ctx, "time://")

	if err != nil {
		t.Fatalf("Failed to invoke ExampleInitializationFunc for time://, %v", err)
	}

	v := t_ex.String()

	now := time.Now()
	year := now.Year()

	str_year := strconv.Itoa(year)

	if !strings.HasPrefix(v, str_year) {
		t.Fatalf("Unexpected value: '%s'", v)
	}
}
