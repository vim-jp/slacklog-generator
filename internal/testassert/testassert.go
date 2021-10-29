package testassert

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

// Equal compare "expect" and "actual". It shows a message with "Fatal" when
// there are some differences.
func Equal(t *testing.T, expect, actual interface{}, msg string, opts ...cmp.Option) {
	t.Helper()
	d := cmp.Diff(expect, actual, opts...)
	if d != "" {
		t.Fatalf("not equal: %s: -want +got\n%s", msg, d)
	}
}
