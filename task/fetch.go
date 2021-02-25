// +build refactor

package task

import (
	"context"

	"github.com/aos-dev/noah/pkg/types"
)

func (t *FetchTask) new() {}
func (t *FetchTask) run(ctx context.Context) error {
	if err := t.GetFetcher().FetchWithContext(ctx, t.GetPath(), t.GetURL()); err != nil {
		return types.NewErrUnhandled(err)
	}
	return nil
}
