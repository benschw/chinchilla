package ep

import (
	"os"
	"os/signal"
	"syscall"
)

func WatchSignals(t chan Trigger) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	signal.Notify(sigCh, syscall.SIGTERM)
	signal.Notify(sigCh, syscall.SIGHUP)

	// main control flow
	for {
		select {

		// If a signal is caught, either shutdown or reload gracefully
		case sig := <-sigCh:
			switch sig {
			case os.Interrupt:
				fallthrough
			case syscall.SIGTERM:
				t <- TriggerStop
				return
			case syscall.SIGHUP:
				t <- TriggerReload
			}
		}
	}

}
