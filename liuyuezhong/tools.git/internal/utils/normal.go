package utils

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func OnExit() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	End:
		for {
			s := <-sigChan
			switch s {
			case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
				break End
			case syscall.SIGHUP:
			default:
			}
		}
		cancel()
	}()

	return ctx
}
