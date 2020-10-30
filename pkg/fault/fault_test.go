package fault

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFault_Pop(t *testing.T) {
	testErr := errors.New("test error")
	count := 3

	f := New()
	assert.Equal(t, 0, len(f.errs), "just init error list len should be 0")

	err := f.Pop()
	assert.Nil(t, err, "error should be nil just after init")
	assert.Equal(t, 0, len(f.errs), "error len should be 0 after pop")

	for i := 0; i < count; i++ {
		f.Append(testErr)
	}
	assert.Equal(t, count, len(f.errs), "error list len should equal after append")

	err = f.Pop()
	assert.NotNil(t, err, "error should not be nil after append errors")
	assert.Equal(t, 0, len(f.errs), "error len should be 0 after pop")

	errList := strings.Split(err.Error(), "\n")
	assert.Equal(t, count, len(errList), "error count should meet")
	assert.Equal(t, testErr.Error(), errList[0], "error msg should meet")
}
