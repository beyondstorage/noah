package task

import (
	"context"
	"crypto/md5"

	"github.com/aos-dev/go-storage/v2/pairs"

	"github.com/qingstor/noah/pkg/types"
)

func (t *MD5SumFileTask) new() {}
func (t *MD5SumFileTask) run(ctx context.Context) error {
	h := md5.New()
	_, err := t.GetStorage().ReadWithContext(
		ctx, t.GetPath(), h, pairs.WithSize(t.GetSize()), pairs.WithOffset(t.GetOffset()))
	if err != nil {
		return types.NewErrUnhandled(err)
	}

	t.SetMD5Sum(h.Sum(nil)[:])
	return nil
}

func (t *MD5SumStreamTask) new() {}
func (t *MD5SumStreamTask) run(_ context.Context) error {
	md5Sum := md5.Sum(t.GetContent().Bytes())
	t.SetMD5Sum(md5Sum[:])
	return nil
}
