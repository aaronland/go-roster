// package roster provides interfaces and methods for defining internal lookup tables for registering
// and instantiatinge custom interfaces with multiple implementations. The expectation is that these
// (roster) interfaces will never be seen by user-facing code.
//
// The generic `Registry[T]` type is the preferred way to build a scheme-based registry; the lower-level
// `Roster` interface remains available for custom needs.
package roster
