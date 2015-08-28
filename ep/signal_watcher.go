package ep

import (
	"os"
	"os/signal"
	"syscall"
)

func NewSignalWatcher() *SignalWatcher {
	w := &SignalWatcher{
		T:  make(chan Trigger),
		ex: make(chan struct{}),
	}
	go watchSignals(w.ex, w.T)
	return w
}

type SignalWatcher struct {
	T  chan Trigger
	ex chan struct{}
}

func (s *SignalWatcher) Stop() {
	close(s.ex)
}

func watchSignals(ex chan struct{}, t chan Trigger) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	signal.Notify(sigCh, syscall.SIGTERM)
	signal.Notify(sigCh, syscall.SIGHUP)

	// main control flow
	for {
		select {
		// when exit chan closes, return
		case <-ex:
			return
		// If a signal is caught, either shutdown or reload gracefully
		case sig := <-sigCh:
			switch sig {
			case os.Interrupt:
				fallthrough
			case syscall.SIGTERM:
				t <- TriggerStop
			case syscall.SIGHUP:
				t <- TriggerReload
			}
		}
	}

}
