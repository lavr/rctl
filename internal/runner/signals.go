package runner

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

// forwardSignals sets up signal forwarding from the parent process to the child.
// Returns a cancel function that should be called when done.
func forwardSignals(cmd *exec.Cmd) func() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	done := make(chan struct{})
	go func() {
		for {
			select {
			case sig := <-sigCh:
				if cmd.Process != nil {
					_ = cmd.Process.Signal(sig)
				}
			case <-done:
				return
			}
		}
	}()

	return func() {
		signal.Stop(sigCh)
		close(done)
	}
}
