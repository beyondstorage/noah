package task

import (
	"testing"

	"github.com/aos-dev/go-storage/v3/types"
	"github.com/stretchr/testify/assert"
)

func Test_calculatePartSize(t *testing.T) {
	var _1m int64 = 1024 * 1024
	var _1g = 1024 * _1m
	var defaultNumber = 10000

	cases := []struct {
		name           string
		numberMax      int
		partSizeMin    int64
		partSizeMax    int64
		length         int64
		expectPartSize int64
		hasErr         bool
	}{
		{
			name:           "5g default part size",
			numberMax:      defaultNumber,
			partSizeMin:    4 * _1m,
			partSizeMax:    5 * _1g,
			length:         5 * _1g,
			expectPartSize: defaultMultipartPartSize,
			hasErr:         false,
		},
		{
			name:           "object no content length",
			numberMax:      defaultNumber,
			partSizeMin:    0,
			partSizeMax:    0,
			length:         0,
			expectPartSize: defaultMultipartPartSize,
			hasErr:         false,
		},
		{
			name:           "object too large",
			numberMax:      100,
			partSizeMin:    0,
			partSizeMax:    4 * _1m,
			length:         5 * _1g,
			expectPartSize: 0,
			hasErr:         true,
		},
		{
			name:           "no restriction",
			numberMax:      0,
			partSizeMin:    0,
			partSizeMax:    0,
			length:         5 * _1g,
			expectPartSize: defaultMultipartPartSize,
			hasErr:         false,
		},
		{
			name:           "no num, only max",
			numberMax:      0,
			partSizeMin:    0,
			partSizeMax:    64 * _1m,
			length:         _1g,
			expectPartSize: 64 * _1m,
			hasErr:         false,
		},
		{
			name:           "no num, only max",
			numberMax:      0,
			partSizeMin:    0,
			partSizeMax:    64 * _1m,
			length:         _1g,
			expectPartSize: 64 * _1m,
			hasErr:         false,
		},
		{
			name:           "no num, only min",
			numberMax:      0,
			partSizeMin:    512 * _1m,
			partSizeMax:    0,
			length:         _1g,
			expectPartSize: 512 * _1m,
			hasErr:         false,
		},
		{
			name:           "only num",
			numberMax:      50,
			partSizeMin:    0,
			partSizeMax:    0,
			length:         10 * _1g,
			expectPartSize: 256 * _1m,
			hasErr:         false,
		},
		{
			name:           "num and min",
			numberMax:      50,
			partSizeMin:    512 * _1m,
			partSizeMax:    0,
			length:         10 * _1g,
			expectPartSize: 512 * _1m,
			hasErr:         false,
		},
		{
			name:           "num, max and min",
			numberMax:      50,
			partSizeMin:    4 * _1m,
			partSizeMax:    250 * _1m,
			length:         10 * _1g,
			expectPartSize: 214748365, // 0.2g + 1
			hasErr:         false,
		},
	}

	for _, tt := range cases {
		obj := types.NewObject(nil, true)
		if tt.numberMax > 0 {
			obj.SetMultipartNumberMaximum(tt.numberMax)
		}
		if tt.partSizeMax > 0 {
			obj.SetMultipartSizeMaximum(tt.partSizeMax)
		}
		if tt.partSizeMin > 0 {
			obj.SetMultipartSizeMinimum(tt.partSizeMin)
		}
		if tt.length > 0 {
			obj.SetContentLength(tt.length)
		}

		if tt.length <= 0 {
			assert.Panics(t, func() {
				_, _ = calculatePartSize(obj)
			}, tt.name)
			continue
		}

		partSize, err := calculatePartSize(obj)

		if tt.hasErr {
			assert.NotNil(t, err, tt.name)
		} else {
			assert.Equal(t, tt.expectPartSize, partSize, tt.name)
			assert.Nil(t, err, tt.name)
		}
	}

}
