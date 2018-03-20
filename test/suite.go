package test

import "testing"

// declare sub test setup func
type SetupSubTest func(t *testing.T) func(t *testing.T)

// empty sub test func
func EmptySubTest() SetupSubTest {
	return func(t *testing.T) func(t *testing.T) { return func(t *testing.T) {} }
}
