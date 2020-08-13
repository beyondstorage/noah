package task

import (
	"context"

	"github.com/Xuanwo/storage/types/pairs"

	"github.com/qingstor/noah/pkg/types"
)

func (t *ReachFileTask) new() {}
func (t *ReachFileTask) run(ctx context.Context) {
	url, err := t.GetReacher().ReachWithContext(ctx, t.GetPath(), pairs.WithExpire(t.GetExpire()))
	if err != nil {
		t.TriggerFault(types.NewErrUnhandled(err))
		return
	}
	t.SetURL(url)
}
