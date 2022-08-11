package hook

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Handler func() error

type hookHandler struct {
	sync.Mutex
	f    []Handler
	done chan os.Signal
}

var hooks hookHandler

// Register adds Handlers to be invoked when the command is terminated.
func Register(handlers ...Handler) {
	hooks.Lock()
	defer hooks.Unlock()
	hooks.f = append(hooks.f, handlers...)
}

// Reset re-instantiate underlying hooks.
func Reset() {
	hooks.Lock()
	defer hooks.Unlock()
	hooks.f = []Handler{}
	hooks.done = make(chan os.Signal, 1)
}

// Listen starts go routine in the background waiting for owning process to be terminated, and when it happens
// it invokes defined Handlers sequentially. Every invocation will reset hooks.
func Listen() {
	Reset()

	signal.Notify(hooks.done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		defer func() {
			signal.Stop(hooks.done)
		}()

		if _, ok := <-hooks.done; !ok {
			// Channel has been closed by calling Close(). Do nothing. Normal termination.
			return
		}

		for _, hook := range hooks.f {
			if err := hook(); err != nil {
				fmt.Printf("failed handling shutdown hook: %s", err.Error())
			}
		}

		os.Exit(130) // INT exit code
	}()
}

// Close closes underlying channel.
func Close() {
	hooks.Lock()
	close(hooks.done)
	hooks.Unlock()
}
