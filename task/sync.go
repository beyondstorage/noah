package task

import (
	"context"

	"github.com/Xuanwo/navvy"
	typ "github.com/Xuanwo/storage/types"

	"github.com/qingstor/noah/pkg/types"
	"github.com/qingstor/noah/utils"
)

func (t *SyncTask) new() {}
func (t *SyncTask) run(ctx context.Context) {
	x := NewListDir(t)
	err := utils.ChooseSourceStorageAsDirLister(x, t)
	if err != nil {
		t.TriggerFault(err)
		return
	}

	if t.GetRecursive() {
		x.SetDirFunc(func(o *typ.Object) {
			sf := NewSync(t)
			sf.SetSourcePath(o.Name)
			sf.SetDestinationPath(o.Name)
			if t.ValidateHandleObjCallback() {
				sf.SetHandleObjCallback(t.GetHandleObjCallback())
			}
			t.GetScheduler().Sync(ctx, sf)
		})
	} else {
		// if not recursive, do nothing with dir
		x.SetDirFunc(func(_ *typ.Object) {})
	}

	var fn []func(task navvy.Task) navvy.Task
	if t.GetIgnoreExisting() {
		fn = append(fn,
			NewIsDestinationObjectNotExistTask,
		)
	}
	if t.GetExisting() {
		fn = append(fn,
			NewIsDestinationObjectExistTask,
		)
	}
	if t.GetUpdate() {
		fn = append(fn,
			NewIsUpdateAtGreaterTask,
		)
	}
	x.SetFileFunc(func(o *typ.Object) {
		sf := NewCopyFile(t)
		sf.SetSourcePath(o.Name)
		sf.SetDestinationPath(o.Name)
		sf.SetCheckTasks(nil)

		// put check tasks outside of copy, to make sure flags' priority is higher than dry-run
		check := NewBetweenStorageCheck(sf)
		sf.GetScheduler().Sync(ctx, check)
		if sf.GetFault().HasError() {
			return
		}
		for _, v := range fn {
			ct := v(check)
			sf.GetScheduler().Sync(ctx, ct)
			if sf.GetFault().HasError() {
				return
			}
			// If any of checks not pass, do not copy this file.
			if result := ct.(types.ResultGetter); !result.GetResult() {
				return
			}
			// If all check passed, we should continue do copy works.
		}

		// if dry-run, only check, and call dry-run func if check passed
		if t.GetDryRun() {
			t.GetDryRunFunc()(o)
			return
		}

		if t.ValidateHandleObjCallback() {
			sf.SetCallbackFunc(func() {
				t.GetHandleObjCallback()(o)
			})
		}
		t.GetScheduler().Async(ctx, sf)
	})
	t.GetScheduler().Sync(ctx, x)
}
