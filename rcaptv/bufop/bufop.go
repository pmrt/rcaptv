package bufop

// Bufop allows to efficiently and repeatedly execute operations on a group of
// numbers by using a window-ed list.
//
//	window: (size=4)
//
// inputs --------10 20 20 30 <- [ 50 60 80 90 ]  <- 80
//
//	^^^^^^^^^^^
//	cache (50+60+80+90)
//
// Next window: [60 80 90 80]
// Next cache: cache - 50 + 80
//
// Operations like Avg() are computed against the cache, which contains the
// current values of the window.
//
// New(`size`) returns a Bufop of window size = `size`. When a number is
// inserted into the buffer only the operations of the entering (+80) and
// leaving (-50) values of the window are needed, caching the rest.
type Bufop struct {
	buf     []float32
	cache   float32
	size    float32
	inserts float32
}

func (b *Bufop) PutInt(n int) (bop *Bufop) {
	l, first := len(b.buf), b.buf[0]
	copy(b.buf[:l-1], b.buf[1:])
	b.buf[l-1] = float32(n)
	b.cache = b.cache + b.buf[l-1] - first

	b.inserts++
	return
}

func (b *Bufop) Avg() float32 {
	return b.cache / min(b.inserts, b.size)
}

func New(size int) *Bufop {
	return &Bufop{
		buf:  make([]float32, size),
		size: float32(size),
	}
}

func min(x, y float32) float32 {
	if x < y {
		return x
	}
	return y
}
