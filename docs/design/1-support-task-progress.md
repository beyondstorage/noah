---
author: Lance Ren <lanceren@yunify.com>
status: draft
updated_at: 2020-02-19
---

# Proposal: Support Task Progress

## Background

For many of our tasks, they cost a lot while running. In case of improving
user's experience, we do need to support return task progress state.
The task caller may use the state for implementing the process bar and so on.

## Proposal

So I propose following changes:

- Add a type `State`, which contains `status description`,
  `task progress remain (or done)` and `task ptogress total`
  in task.
- Init a global progress center for every registered task.
  and each task reports its `State` into the progress center
  every 500ms (configurable).
- Add a unify component to combine all tasks' state and compose
  the output.

After all these work, we can work well with progress bar:

```go
package main

import (
    "log"
    "time"

	"github.com/aos-dev/noah/task"
    "github.com/schollz/progressbar/v2"
)

func main()  {
 	// Init a task.
 	t := task.NewCopyFileTask(rootTask)
    t.SetCheckTasks(nil)

    closeSig := make(chan struct{})
 	defer func() {
        close(closeSig)
    }()

    go func() {
        state := <-t.GetProgressState()
        bar := progressbar.New64(state.GetTotal())
        for {
            select{
            case s := <-t.GetProgressState():
                bar.Set(s.GetDone())
            case <-closeSig:
                return
            }
        }
        bar.Finish()
    }()
    t.Run()
}
```

## Rationale

A task with progress is the most fluent, natural way in implement.
And noah was designed to be a task-driven frame, so I think it's
proper to return progress state while task is running.

### Why not get task total at very beginning

We cannot handle sub-task count when init a task. Because task-list
is implemented as async task, and we can't get the total from object
storage services.

### Why not set callback func between task and its sub-task

Many tasks were submitted asynchronously, so if we want to set callback
func between tasks, we may need to add a lot of mutex to avoid data race.
It will cost too much in performance.

## Compatibility

No breaking changes.

## Implementation

Most of the work would be done by the author of this proposal.
