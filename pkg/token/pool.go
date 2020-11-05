package token

// Pool handle a reusable token pool
type Pool struct {
	ch chan struct{}
}

// Take implement Tokener.Take
func (p *Pool) Take() {
	<-p.ch
}

// Return implement Tokener.Return
func (p *Pool) Return() {
	p.ch <- struct{}{}
}

// Close implement Tokener.Close
func (p *Pool) Close() {
	close(p.ch)
}

// NewPool create a Pool with given cap and return its pointer
func NewPool(cap int) *Pool {
	if cap <= 0 {
		cap = 10
	}
	ch := make(chan struct{}, cap)
	for i := 0; i < cap; i++ {
		ch <- struct{}{}
	}
	return &Pool{ch: ch}
}
