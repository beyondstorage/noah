package task

import (
	"bytes"
	"context"
	"io"
	"testing"

	typ "github.com/aos-dev/go-storage/v2/types"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/qingstor/noah/pkg/fault"
	"github.com/qingstor/noah/pkg/mock"
)

func TestMD5SumFileTask_run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	store := mock.NewMockStorager(ctrl)
	srcReader := mock.NewMockReadCloser(ctrl)
	srcPath := uuid.New().String()
	size := int64(1024)

	task := MD5SumFileTask{}
	task.SetFault(fault.New())
	task.SetStorage(store)
	task.SetPath(srcPath)
	task.SetSize(size)
	task.SetOffset(0)

	srcReader.EXPECT().Close().Do(func() {})
	store.EXPECT().ReadWithContext(gomock.Eq(ctx), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, path string, pairs ...*typ.Pair) (r io.ReadCloser, err error) {
			assert.Equal(t, srcPath, path)
			return srcReader, nil
		})
	srcReader.EXPECT().Read(gomock.Any()).DoAndReturn(func(p []byte) (n int, err error) {
		return 768, io.EOF
	})

	task.run(ctx)
	assert.NotEmpty(t, task.GetMD5Sum())
	assert.Empty(t, task.GetFault().Error())
}

func TestMD5SuSteamTask_run(t *testing.T) {
	task := MD5SumStreamTask{}
	task.SetFault(fault.New())
	task.SetContent(&bytes.Buffer{})

	task.run(context.Background())
	assert.NotEmpty(t, task.GetMD5Sum())
	assert.Empty(t, task.GetFault().Error())
}
