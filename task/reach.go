package task

import (
	"context"

	"github.com/aos-dev/go-storage/v2/pairs"

	"github.com/aos-dev/noah/pkg/types"
)

func (t *ReachFileTask) new() {}
func (t *ReachFileTask) run(ctx context.Context) error {
	url, err := t.GetReacher().ReachWithContext(ctx, t.GetPath(), pairs.WithExpire(t.GetExpire()))
	if err != nil {
		return types.NewErrUnhandled(err)
	}
	t.SetURL(url)
	return nil
}
