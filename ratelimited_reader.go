package copy

import (
	"io"

	"go.uber.org/ratelimit"
)

type RateLimitedReader struct {
	src     io.Reader
	limiter ratelimit.Limiter
}

// NewRateLimitedReader
// n means the number of KB to be read per second
func NewRateLimitedReader(src io.Reader, n int64) io.Reader {
	return &RateLimitedReader{
		src:     src,
		limiter: ratelimit.New(int(n)),
	}
}

func (lr *RateLimitedReader) Read(p []byte) (n int, err error) {
	n, e := lr.src.Read(p)
	if e != nil && e != io.EOF {
		return n, e
	}
	if n > 0 {
		nkb := n / 1024
		if nkb == 0 {
			nkb = 1
		}
		for i := 0; i < nkb; i++ {
			lr.limiter.Take()
		}
	}
	return n, e
}
