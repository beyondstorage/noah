package token

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewPool(t *testing.T) {
	t.Run("cap 10", func(t *testing.T) {
		c := 10
		p := NewPool(c)
		assert.NotPanics(t, func() {
			for i := 0; i < c; i++ {
				p.Take()
			}
			p.Close()
		})
	})

	t.Run("negative cap", func(t *testing.T) {
		p := NewPool(-1)
		assert.NotPanics(t, func() {
			for i := 0; i < defaultCap; i++ {
				p.Take()
			}
			p.Close()
		})
	})

	t.Run("more take with return", func(t *testing.T) {
		c := 3
		p := NewPool(c)
		go func() {
			for i := 0; i < c; i++ {
				time.Sleep(time.Second)
				p.Return()
			}
		}()

		assert.NotPanics(t, func() {
			for i := 0; i < 2*c; i++ {
				p.Take()
			}
			p.Close()
		})
	})

	t.Run("panic with double close", func(t *testing.T) {
		p := NewPool(1)
		p.Close()
		assert.Panics(t, func() {
			p.Close()
		})
	})

	t.Run("more take after close", func(t *testing.T) {
		p := NewPool(1)
		p.Close()
		assert.NotPanics(t, func() {
			for i := 0; i < 10; i++ {
				p.Take()
			}
		})
	})
}
