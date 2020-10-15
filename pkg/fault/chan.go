package fault

import "sync"

// Syncer is the struct to handle error channel and do sync job
type Syncer struct {
	sync.Once
	errChan  chan error
	exitChan chan struct{}
}

// NewSyncer conduct and return a *Syncer
func NewSyncer() *Syncer {
	sc := &Syncer{}
	sc.errChan = make(chan error, 1)
	sc.exitChan = make(chan struct{})
	return sc
}

// GetErrChan return the errChan to transfer err
func (sc *Syncer) GetErrChan() chan error {
	return sc.errChan
}

// Wait will hang until get the exit signal
func (sc *Syncer) Wait() {
	<-sc.exitChan
}

// Finish close the errChan and send the exit signal to exit syncer
// use once to ensure only close once
func (sc *Syncer) Finish() {
	sc.Once.Do(func() {
		close(sc.errChan)
		sc.exitChan <- struct{}{}
		close(sc.exitChan)
	})
}
