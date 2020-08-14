package utils

import (
	"fmt"

	"github.com/aos-dev/go-storage/v2"

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

// ChooseSourceStorageAsDirLister will choose the source storage to fill as the dir lister.
func ChooseSourceStorageAsDirLister(x interface {
	types.PathSetter
	types.DirListerSetter
}, y interface {
	types.SourcePathGetter
	types.SourceStorageGetter
}) (err error) {
	x.SetPath(y.GetSourcePath())

	lister, ok := y.GetSourceStorage().(storage.DirLister)
	if !ok {
		return types.NewErrStorageInsufficientAbility(nil)
	}
	x.SetDirLister(lister)
	return
}

// ChooseStorageAsDirLister will choose the storage to fill as the dir lister.
func ChooseStorageAsDirLister(x interface {
	types.PathSetter
	types.DirListerSetter
}, y interface {
	types.PathGetter
	types.StorageGetter
}) (err error) {
	x.SetPath(y.GetPath())

	lister, ok := y.GetStorage().(storage.DirLister)
	if !ok {
		return types.NewErrStorageInsufficientAbility(nil)
	}
	x.SetDirLister(lister)
	return
}

// ChooseStorageAsPrefixLister will choose the storage to fill as the prefix lister.
func ChooseStorageAsPrefixLister(x interface {
	types.PathSetter
	types.PrefixListerSetter
}, y interface {
	types.PathGetter
	types.StorageGetter
}) (err error) {
	x.SetPath(y.GetPath())

	lister, ok := y.GetStorage().(storage.PrefixLister)
	if !ok {
		return types.NewErrStorageInsufficientAbility(nil)
	}
	x.SetPrefixLister(lister)
	return
}

// ChooseDestinationStorageAsIndexSegmenter will choose the destination storage as a segmenter.
func ChooseDestinationStorageAsIndexSegmenter(x interface {
	types.PathSetter
	types.IndexSegmenterSetter
}, y interface {
	types.DestinationPathGetter
	types.DestinationStorageGetter
}) (err error) {
	x.SetPath(y.GetDestinationPath())

	segmenter, ok := y.GetDestinationStorage().(storage.IndexSegmenter)
	if !ok {
		return types.NewErrStorageInsufficientAbility(nil)
	}
	x.SetIndexSegmenter(segmenter)
	return
}

// ChooseDestinationIndexSegmenter will choose the destination storage as a segmenter.
func ChooseDestinationIndexSegmenter(x interface {
	types.DestinationPathSetter
	types.DestinationIndexSegmenterSetter
}, y interface {
	types.DestinationPathGetter
	types.DestinationStorageGetter
}) (err error) {
	x.SetDestinationPath(y.GetDestinationPath())

	segmenter, ok := y.GetDestinationStorage().(storage.IndexSegmenter)
	if !ok {
		return types.NewErrStorageInsufficientAbility(nil)
	}
	x.SetDestinationIndexSegmenter(segmenter)
	return
}
