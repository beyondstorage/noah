package token

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlankToken(t *testing.T) {
	b := BlankToken{}
	assert.NotPanics(t, func() {
		b.Take()
		b.Return()
		b.Close()
	})
}

func TestContextWithTokener(t *testing.T) {
	p := NewPool(1)
	t.Run("nil context", func(t *testing.T) {
		ctx := ContextWithTokener(nil, p)
		tk := FromContext(ctx)
		assert.Equal(t, p, tk)
	})

	t.Run("nil tokener", func(t *testing.T) {
		ctx := ContextWithTokener(context.Background(), nil)
		assert.Nil(t, ctx.Value(tokenKey))
	})

	t.Run("normal context with multiple set", func(t *testing.T) {
		ctx := ContextWithTokener(context.Background(), nil)
		tk := FromContext(ctx)
		_, ok := tk.(BlankToken)
		assert.True(t, ok)

		ctx = ContextWithTokener(nil, p)
		tk = FromContext(ctx)
		assert.Equal(t, p, tk)

		b := BlankToken{}
		ctx = ContextWithTokener(ctx, b)
		tk = FromContext(ctx)
		assert.Equal(t, b, tk)
	})

}

func TestFromContext(t *testing.T) {
	t.Run("nil context", func(t *testing.T) {
		tk := FromContext(nil)
		_, ok := tk.(BlankToken)
		assert.True(t, ok)
	})

	t.Run("tokener not set", func(t *testing.T) {
		tk := FromContext(context.Background())
		_, ok := tk.(BlankToken)
		assert.True(t, ok)
	})

	t.Run("tokener set", func(t *testing.T) {
		p := NewPool(1)
		ctx := ContextWithTokener(context.Background(), p)
		tk := FromContext(ctx)
		assert.Equal(t, p, tk)
	})
}
