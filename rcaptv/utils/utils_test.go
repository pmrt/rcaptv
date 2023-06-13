package utils

import (
	"testing"
)

func TestPrependHash(t *testing.T) {
	prefix, hash := "sha256=", "efff62e8394965726992ca425ac5aa9550b4e524e98b936b6bdddc2e86d53990"
	b1 := make([]byte, 0, 64+7)
	b1 = append(b1, []byte(hash)...)

	b1 = Prepend(b1, []byte(prefix))

	got, want := string(b1), string([]byte(prefix+hash))
	if got != want {
		t.Fatalf("\n  got: %+v\n want: %+v", b1, []byte(prefix+hash))
	}
}

func TestPrependWithoutAllocations(t *testing.T) {
	const (
		n      = 100
		prefix = "prefix"
		buf    = "abcd"
	)

	// a buffer with enough capacity should not allocate during Prepend
	b1 := make([]byte, 0, len(buf)+len(prefix))
	b1 = append(b1, []byte(buf)...)
	b2 := []byte(prefix)

	if avg := testing.AllocsPerRun(n, func() {
		Prepend(b1, b2)
	}); avg != 0 {
		t.Fatal("expected prepend to make no allocations")
	}
}
