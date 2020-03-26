package task

import (
	"github.com/Xuanwo/navvy"
	typ "github.com/Xuanwo/storage/types"

	"github.com/qingstor/noah/pkg/types"
	"github.com/qingstor/noah/utils"
)

func (t *SyncTask) new() {}
func (t *SyncTask) run() {
	x := NewListDir(t)
	utils.ChooseSourceStorage(x, t)
	if t.GetRecursive() {
		x.SetDirFunc(func(o *typ.Object) {
			sf := NewSync(t)
			sf.SetSourcePath(o.Name)
			sf.SetDestinationPath(o.Name)
			t.GetScheduler().Sync(sf)
		})
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
		sf.SetCheckTasks(fn)

		// if dry-run, only check, and call dry-run func if check passed
		if t.GetDryRun() {
			check := NewBetweenStorageCheck(sf)
			sf.GetScheduler().Sync(check)
			if sf.GetFault().HasError() {
				return
			}
			for _, v := range fn {
				ct := v(check)
				sf.GetScheduler().Sync(ct)
				// If any of checks not pass, do not copy this file.
				if result := ct.(types.ResultGetter); !result.GetResult() {
					return
				}
				// If all check passed, we should continue do copy works.
			}
			t.GetDryRunFunc()(o)
			return
		}
		t.GetScheduler().Async(sf)
	})
	t.GetScheduler().Sync(x)
}
