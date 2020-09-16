package task

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/aos-dev/go-storage/v2/services"
	typ "github.com/aos-dev/go-storage/v2/types"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/qingstor/noah/pkg/fault"
	"github.com/qingstor/noah/pkg/mock"
)

func TestBetweenStorageCheckTask_run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	cases := []struct {
		name         string
		expectObject *typ.Object
		expectErr    error
	}{
		{
			"normal",
			&typ.Object{},
			nil,
		},
		{
			"dst object not exist",
			nil,
			services.ErrObjectNotExist,
		},
		{
			"error",
			nil,
			errors.New("unhandled error"),
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			srcStore := mock.NewMockStorager(ctrl)
			dstStore := mock.NewMockStorager(ctrl)
			srcPath := uuid.New().String()
			dstPath := uuid.New().String()

			task := BetweenStorageCheckTask{}
			task.SetSourceStorage(srcStore)
			task.SetDestinationStorage(dstStore)
			task.SetSourcePath(srcPath)
			task.SetDestinationPath(dstPath)
			task.SetFault(fault.New())

			srcStore.EXPECT().StatWithContext(gomock.Eq(ctx), gomock.Any()).
				DoAndReturn(func(ctx context.Context, path string, pairs ...*typ.Pair) (o *typ.Object, err error) {
					assert.Equal(t, srcPath, path)
					return &typ.Object{Name: srcPath}, nil
				})
			srcStore.EXPECT().String().DoAndReturn(func() string {
				return "src"
			}).AnyTimes()
			dstStore.EXPECT().StatWithContext(gomock.Eq(ctx), gomock.Any()).
				DoAndReturn(func(ctx context.Context, path string, pairs ...*typ.Pair) (o *typ.Object, err error) {
					assert.Equal(t, dstPath, path)
					return tt.expectObject, tt.expectErr
				})
			dstStore.EXPECT().String().DoAndReturn(func() string {
				return "dst"
			}).AnyTimes()

			task.run(ctx)

			assert.NotNil(t, task.GetSourceObject())
			if tt.expectObject != nil {
				assert.NotNil(t, task.GetDestinationObject())
			} else {
				if tt.expectErr == services.ErrObjectNotExist {
					assert.Nil(t, task.GetDestinationObject())
				} else {
					assert.Panics(t, func() {
						task.GetDestinationObject()
					})
				}
			}
		})
	}
}

func TestIsDestinationObjectExistTask_run(t *testing.T) {
	ctx := context.Background()

	t.Run("destination object not exist", func(t *testing.T) {
		task := IsDestinationObjectExistTask{}
		task.SetDestinationObject(nil)

		task.run(ctx)

		assert.Equal(t, false, task.GetResult())
	})

	t.Run("destination object exists", func(t *testing.T) {
		task := IsDestinationObjectExistTask{}
		task.SetDestinationObject(&typ.Object{})

		task.run(ctx)

		assert.Equal(t, true, task.GetResult())
	})
}

func TestIsSizeEqualTask_run(t *testing.T) {
	ctx := context.Background()

	t.Run("size equal", func(t *testing.T) {
		task := IsSizeEqualTask{}
		task.SetSourceObject(&typ.Object{Size: 111})
		task.SetDestinationObject(&typ.Object{Size: 111})

		task.run(ctx)

		assert.Equal(t, true, task.GetResult())
	})

	t.Run("size not equal", func(t *testing.T) {
		task := IsSizeEqualTask{}
		task.SetSourceObject(&typ.Object{Size: 222})
		task.SetDestinationObject(&typ.Object{Size: 111})

		task.run(ctx)

		assert.Equal(t, false, task.GetResult())
	})
}

func TestIsUpdateAtGreaterTask_run(t *testing.T) {
	ctx := context.Background()

	t.Run("updated at greater", func(t *testing.T) {
		task := IsUpdateAtGreaterTask{}
		task.SetSourceObject(&typ.Object{UpdatedAt: time.Now().Add(time.Hour)})
		task.SetDestinationObject(&typ.Object{UpdatedAt: time.Now()})

		task.run(ctx)

		assert.Equal(t, true, task.GetResult())
	})

	t.Run("updated at not greater", func(t *testing.T) {
		task := IsUpdateAtGreaterTask{}
		task.SetSourceObject(&typ.Object{UpdatedAt: time.Now()})
		task.SetDestinationObject(&typ.Object{UpdatedAt: time.Now().Add(time.Hour)})

		task.run(ctx)

		assert.Equal(t, false, task.GetResult())
	})
}

func TestIsSourcePathExcludeIncludeTask_run(t *testing.T) {
	ctx := context.Background()
	reg := regexp.MustCompile("-")
	nonMatchReg := regexp.MustCompile("^$")
	path := uuid.New().String()

	tests := []struct {
		name       string
		sourcePath string
		excludeReg *regexp.Regexp
		includeReg *regexp.Regexp
		wantRes    bool
	}{
		{
			name:       "exclude not set",
			sourcePath: path,
			excludeReg: nil,
			includeReg: nil,
			wantRes:    true,
		},
		{
			name:       "exclude set non-match",
			sourcePath: path,
			excludeReg: nonMatchReg,
			includeReg: nil,
			wantRes:    true,
		},
		{
			name:       "exclude set match, include not set",
			sourcePath: path,
			excludeReg: reg,
			includeReg: nil,
			wantRes:    false,
		},
		{
			name:       "exclude set match, include set non-match",
			sourcePath: path,
			excludeReg: reg,
			includeReg: nonMatchReg,
			wantRes:    false,
		},
		{
			name:       "exclude set match, include set match",
			sourcePath: path,
			excludeReg: reg,
			includeReg: reg,
			wantRes:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := IsSourcePathExcludeIncludeTask{}
			task.SetSourcePath(tt.sourcePath)
			task.SetExcludeRegexp(tt.excludeReg)
			task.SetIncludeRegexp(tt.includeReg)
			task.run(ctx)

			assert.Equal(t, tt.wantRes, task.GetResult())
		})
	}
}
