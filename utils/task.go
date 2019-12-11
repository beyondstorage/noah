package utils

import (
	"fmt"

	"github.com/Xuanwo/storage"
	"github.com/qingstor/noah/constants"
	"github.com/qingstor/noah/pkg/types"
)

// CalculatePartSize will calculate the object's part size.
func CalculatePartSize(size int64) (partSize int64, err error) {
	partSize = constants.DefaultPartSize

	if size > constants.MaximumObjectSize {
		return 0, fmt.Errorf("calculate part size failed: {%w}", types.NewErrLocalFileTooLarge(nil, size))
	}

	for size/partSize >= int64(constants.MaximumMultipartNumber) {
		if partSize < constants.MaximumAutoMultipartSize {
			partSize = partSize << 1
			continue
		}
		// Try to adjust partSize if it is too small and account for
		// integer division truncation.
		partSize = size/int64(constants.MaximumMultipartNumber) + 1
		break
	}
	return
}

// ChooseDestinationStorage will choose the destination storage to fill.
func ChooseDestinationStorage(x interface {
	types.PathSetter
	types.StorageSetter
}, y interface {
	types.DestinationPathGetter
	types.DestinationStorageGetter
}) {
	x.SetPath(y.GetDestinationPath())
	x.SetStorage(y.GetDestinationStorage())
}

// ChooseSourceStorage will choose the source storage to fill.
func ChooseSourceStorage(x interface {
	types.PathSetter
	types.StorageSetter
}, y interface {
	types.SourcePathGetter
	types.SourceStorageGetter
}) {
	x.SetPath(y.GetSourcePath())
	x.SetStorage(y.GetSourceStorage())
}

// ChooseDestinationStorageAsSegmenter will choose the destination storage as a segmenter.
func ChooseDestinationStorageAsSegmenter(x interface {
	types.PathSetter
	types.SegmenterSetter
}, y interface {
	types.DestinationPathGetter
	types.DestinationStorageGetter
}) (err error) {
	x.SetPath(y.GetDestinationPath())

	segmenter, ok := y.GetDestinationStorage().(storage.Segmenter)
	if !ok {
		return types.NewErrStorageInsufficientAbility(nil)
	}
	x.SetSegmenter(segmenter)
	return
}

// ChooseDestinationSegmenter will choose the destination storage as a segmenter.
func ChooseDestinationSegmenter(x interface {
	types.DestinationPathSetter
	types.DestinationSegmenterSetter
}, y interface {
	types.DestinationPathGetter
	types.DestinationStorageGetter
}) (err error) {
	x.SetDestinationPath(y.GetDestinationPath())

	segmenter, ok := y.GetDestinationStorage().(storage.Segmenter)
	if !ok {
		return types.NewErrStorageInsufficientAbility(nil)
	}
	x.SetDestinationSegmenter(segmenter)
	return
}
