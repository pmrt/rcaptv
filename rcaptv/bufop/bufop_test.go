package bufop

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAvg(t *testing.T) {
	t.Parallel()
	type expect struct {
		buf    []float32
		avg    float32
		cache  float32
		lenf32 float32
	}

	table := []expect{
		{buf: []float32{0, 0, 0, 10}, avg: 10, cache: 10},
		{buf: []float32{0, 0, 10, 20}, avg: 15, cache: 30},
		{buf: []float32{0, 10, 20, 30}, avg: 20, cache: 60},
		{buf: []float32{10, 20, 30, 30}, avg: 22.5, cache: 90},
		{buf: []float32{20, 30, 30, 40}, avg: 30, cache: 120},
		{buf: []float32{30, 30, 40, 80}, avg: 45, cache: 180},
		{buf: []float32{30, 40, 80, 5}, avg: 38.75, cache: 155},
		{buf: []float32{40, 80, 5, 6}, avg: 32.75, cache: 131},
		{buf: []float32{80, 5, 6, 2}, avg: 23.25, cache: 93},
		{buf: []float32{5, 6, 2, 5}, avg: 4.5, cache: 18},
	}
	input := []int{10, 20, 30, 30, 40, 80, 5, 6, 2, 5}

	size := 4
	b := New(size)
	for i, row := range table {
		b.PutInt(input[i])
		{
			got, want := b.buf, row.buf
			if !cmp.Equal(got, want) {
				t.Fatalf("buf: got %+v, want %+v", got, want)
			}
		}

		{
			got, want := b.cache, row.cache
			if !cmp.Equal(got, want) {
				t.Fatalf("cache: got %+v, want %+v", got, want)
			}
		}

		{
			got, want := b.Avg(), row.avg
			if !cmp.Equal(got, want) {
				t.Fatalf("avg: got %+v, want %+v", got, want)
			}
		}
	}
}
