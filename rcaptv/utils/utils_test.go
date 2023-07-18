package utils

import (
	"math"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-test/deep"
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
	if Abs(int(math.Inf(-1))) != 9223372036854775807 {
		t.Fatal("expected abs(Inf(-1)) to be 0")
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		a, b int
		want int
	}{
		{a: 6, b: 5, want: 5},
		{a: 10, b: 5, want: 5},
		{a: 0, b: 5, want: 0},
		{a: 4, b: 5, want: 4},
		{a: -1, b: 5, want: 1},
		{a: -1, b: 0, want: 0},
		{a: 0, b: int(math.Inf(1)), want: 0},
		{a: 0, b: int(math.Inf(-1)), want: 0},
		{a: -0, b: 0, want: 0},
		{a: 4, b: 4, want: 4},
		{a: -6, b: 5, want: 5},
		{a: -9223372036854775807, b: 5, want: 5},
		{a: -9223372036854775808, b: 5, want: 5},
		{a: 9223372036854775807, b: 5, want: 5},
	}
	for i, test := range tests {
		got := Min(test.a, test.b)
		want := test.want
		if got != want {
			t.Fatalf("wrong min, test#%d got:%d want:%d", i+1, got, want)
		}
	}
}

func TestRemoveKey(t *testing.T) {
	tests := []struct {
		input  []string
		target string
		want   []string
	}{
		{
			input:  []string{"aa", "bb", "cc", "dd"},
			target: "bb",
			want:   []string{"aa", "cc", "dd"},
		},
		{
			input:  []string{"aa", "bb", "cc", "dd"},
			target: "aa",
			want:   []string{"bb", "cc", "dd"},
		},
		{
			input:  []string{"aa", "bb", "cc", "dd"},
			target: "dd",
			want:   []string{"aa", "bb", "cc"},
		},
	}

	for _, test := range tests {
		got := RemoveKey(test.input, test.target)
		want := test.want
		t.Logf("got:\n%s\n", spew.Sdump(got))
		t.Logf("want:\n%s\n", spew.Sdump(want))
		t.Logf("test.input:\n%s\n", spew.Sdump(test.input))
		if diff := deep.Equal(got, want); diff != nil {
			t.Fatal(diff)
		}
	}
}
