package utils

import (
	"testing"
)

func TestPrependHash(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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

func TestTruncateSecret(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		n     int
		want  string
	}{
		{input: "123456", n: 3, want: "123XXX"},
		{input: "ABCDEFGHJK", n: 2, want: "ABXXXXXXXX"},
		{input: "QWERTYPASSWORD", n: 3, want: "QWEXXXXXXXXXXX"},
	}
	for _, test := range tests {
		got := TruncateSecret(test.input, test.n)
		want := test.want
		if got != want {
			t.Fatalf("wrong truncation, got:%s want:%s", got, want)
		}
	}
}

func TestAbs(t *testing.T) {
	t.Parallel()
	if Abs(100) != 100 {
		t.Fatal("expected abs(100) to be 100")
	}
	if Abs(-100) != 100 {
		t.Fatal("expected abs(-100) to be 100")
	}
	if Abs(142340) != 142340 {
		t.Fatal("expected abs(142340) to be 142340")
	}
	if Abs(-142340) != 142340 {
		t.Fatal("expected abs(-142340) to be 142340")
	}
	if Abs(0) != 0 {
		t.Fatal("expected abs(0) to be 0")
	}
}
