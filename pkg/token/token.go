package token

import "context"

// Tokener is an interface to handle token
type Tokener interface {
	// Take a token from Tokener
	Take()
	// Return a token into Tokener
	Return()
	// Close Tokener
	Close()
}

// BlankToken do nothing, just implement Tokener interface
type BlankToken struct{}

// Take implement Tokener.Take
func (b BlankToken) Take() {}

// Return implement Tokener.Return
func (b BlankToken) Return() {}

// Close implement Tokener.Close
func (b BlankToken) Close() {}

type contextKey struct{}

// tokenKey is used as key to store Tokener in context
var tokenKey contextKey

// ContextWithTokener set Tokener into given context and return
func ContextWithTokener(ctx context.Context, tk Tokener) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	// if nil Tokener was given, return ctx directly
	if tk == nil {
		return ctx
	}

	return context.WithValue(ctx, tokenKey, tk)
}

// FromContext get Tokener from context
// Notice: If ctx is nil or no Tokener was set before, it will return BlankToken
func FromContext(ctx context.Context) Tokener {
	if ctx == nil {
		return BlankToken{}
	}
	tk, ok := ctx.Value(tokenKey).(Tokener)
	if !ok {
		return BlankToken{}
	}
	return tk
}
