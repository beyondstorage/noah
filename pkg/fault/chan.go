package fault

import "sync"

type Syncer struct {
	sync.Once
	errChan  chan error
	exitChan chan struct{}
}

func NewSyncer() *Syncer {
	sc := &Syncer{}
	sc.errChan = make(chan error, 1)
	sc.exitChan = make(chan struct{})
	return sc
}

func (sc *Syncer) GetErrChan() chan error {
	return sc.errChan
}

func (sc *Syncer) Wait() {
	<-sc.exitChan
}

func (sc *Syncer) Finish() {
	sc.Once.Do(func() {
		close(sc.exitChan)
	})
}
