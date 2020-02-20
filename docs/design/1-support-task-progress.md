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

- Add a type (maybe like: progress by percent) in time-cost task,
  so that we can get progress now.
- Set progress state asynchronously in the progress of task running.

After all these work, we can work well with progress bar:

```go
package main

import (
    "log"
    "time"

	"github.com/qingstor/noah/task"
    "github.com/schollz/progressbar/v2"
)

func main()  {
 	// Init a task.
 	t := task.NewCopyFileTask(rootTask)
    t.SetCheckTasks(nil)

    closeSig := make(chan struct{})
 	bar := progressbar.New(16 * 1024 * 1024)
 	defer func() {
        close(closeSig)
 	    bar.Finish()
    }()

    go func() {
        for {
            select{
            case <-time.Tick(time.Second):
                bar.Set(t.GetProgressState())
            case <-closeSig:
                return
            }
        }
    }()
    t.Run()
}
```

## Rationale

A task with progress is the most fluent, natural way in implement.
And noah was designed to be a task-driven frame, so I think it's
proper to return progress state while task is running.

There is another question: How to define the progress?
Firstly, it is easy to treat file-transfer-type task's total size
divided by read bytes as the progress. However, what if other type
tasks? Such as remove a dir or sync a dir?
So, how about treating the whole, including sub-task, giant task as
one, and the total task divided by task-done count as the progress?
In this way, we could unify the stream transfer task, file transfer
task and other time-consuming task. All we need to do is, get the
count of total sub-task, add after each task done, and finish the
task as usual.

## Compatibility

No breaking changes.

## Implementation

Most of the work would be done by the author of this proposal.
