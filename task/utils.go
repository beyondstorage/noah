package task

import (
	"fmt"

	"github.com/aos-dev/go-storage/v3/types"
)

func calculatePartSize(o *types.Object) (int64, error) {
	maxNum, numOK := o.GetMultipartNumberMaximum()
	maxSize, maxOK := o.GetMultipartSizeMaximum()
	minSize, minOK := o.GetMultipartSizeMinimum()

	partSize := defaultMultipartPartSize
	objSize := o.MustGetContentLength()

	// assert object size too large
	if maxOK && numOK {
		if objSize > maxSize*int64(maxNum) {
			return 0, fmt.Errorf("calculate part size failed: {object size too large}")
		}
	}

	// if multipart number no restrict, just check max and min multipart part size
	if !numOK {
		for {
			if maxOK && partSize > maxSize {
				partSize = partSize >> 1
				continue
			}
			if minOK && partSize < minSize {
				partSize = partSize << 1
				continue
			}
			return partSize, nil
		}
	}

	// if multipart number has maximum restriction, count part size dynamically
	for {
		// if partSize larger than maximum, double part size
		if objSize/partSize >= int64(maxNum) {
			// objSize > maxSize * maxNum has been asserted before,
			// so we do not need check
			partSize = partSize << 1
			continue
		}

		// if partSize less than minimum, double part size
		if minOK && partSize < minSize {
			partSize = partSize << 1
			continue
		}

		// if partSize too large, try to count by the max part number
		if maxOK && partSize > maxSize {
			// Try to adjust partSize if it is too small and account for
			// integer division truncation.
			partSize = objSize/int64(maxNum) + 1
			return partSize, nil
		}

		// otherwise, use the part size, which is not too large or too small
		break
	}
	return partSize, nil
}

// validatePartSize used to check user-input part size
func validatePartSize(o *types.Object, partSize int64) error {
	if min, ok := o.GetMultipartSizeMinimum(); ok && partSize < min {
		return fmt.Errorf("part size must larger than {%d}", min)
	}
	if max, ok := o.GetMultipartSizeMaximum(); ok && partSize > max {
		return fmt.Errorf("part size must less than {%d}", max)
	}
	num, ok := o.GetMultipartNumberMaximum()
	if ok {
		parts := o.MustGetContentLength() / partSize
		if parts > int64(num) {
			return fmt.Errorf("parts count at part size <%d> "+
				"will be larger than max number {%d}", partSize, num)
		}
	}
	return nil
}
