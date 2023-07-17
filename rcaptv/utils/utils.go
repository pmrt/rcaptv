package utils

import (
	"math"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"
	"unsafe"
)

// Remove takes an slice and an index `i` and returns the same slice without
// slice[i]. The original slice will be mutated, don't use it.
func Remove[T any](slice []T, i int) []T {
	if i < 0 {
		return slice
	}
	copy(slice[i:], slice[i+1:])
	return slice[:len(slice)-1]
}

func Find(s []string, key string) int {
	for i, n := range s {
		if n == key {
			return i
		}
	}
	return -1
}

// remove takes an slice and a key and returns the same slice without the
// key element. The original slice will be mutated, don't use it.
func RemoveKey(s []string, key string) []string {
	return Remove(s, Find(s, key))
}

func WaitInterrupt() os.Signal {
	sigint := make(chan os.Signal, 1)
	signal.Notify(
		sigint,
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	return <-sigint
}

func StrPtr(s string) *string {
	return &s
}

func StringToByte(s string) (b []byte) {
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh.Data = sh.Data
	bh.Cap = sh.Len
	bh.Len = sh.Len
	return b
}

// Prepend uses append and copy to make the inverse operation of append,
// returning an slice with src before dst, with as few allocations as possible.
//
// If `cap(dst) >= cap(src) + cap(dst)` prepend does not allocate.
//
// Returns `dst` with the new length, so use it with `a = prepend(a, b)`.
// Otherwise with just `prepend(a, b)` a will have the old length.
func Prepend(dst []byte, src []byte) []byte {
	l := len(src)
	// Add as many empty 0 to dst as src len
	for i := 0; i < l; i++ {
		// If there is spare capacity append extends dst length, otherwise it
		// allocates
		dst = append(dst, 0)
	}
	// copy dst to the second half. Note: dst[:] = dst[:len(dst)]
	copy(dst[l:], dst[:])
	// copy src to the first half
	copy(dst[:l], src)
	// return dst with the new length
	return dst
}

func TruncateSecret(s string, n int) string {
	return s[:n] + strings.Repeat("X", len(s)-n)
}

func Abs(x int) int {
	if x < 0 {
		if x == int(math.Inf(-1)) {
			// if x = int(math.Inf(-1)) = -9223372036854775808, to prevent a negative number
			// we cause an overflow and turn math.Inf(-1) into math.Inf(1)-1. It's not perfect
			// because we Inf is bigger than math.Inf(1)-1 but probably ok for most cases
			return x - 1
		}
		return ^x + 1
	}
	return x
}

func Min(x, y int) int {
	x, y = Abs(x), Abs(y)
	if x < y {
		return x
	}
	return y
}
