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
	f       []Handler
	signals chan os.Signal
	open    bool
}

var hooks hookHandler

// Register adds Handlers to be invoked when the command is terminated.
func Register(handlers ...Handler) {
	hooks.Lock()
	hooks.f = append(hooks.f, handlers...)
	hooks.Unlock()
}

// Reset re-instantiate underlying hooks.
func Reset() {
	hooks.Lock()
	hooks.open = false
	hooks.f = []Handler{}
	hooks.signals = make(chan os.Signal, 1)
	hooks.Unlock()
}

// Listen starts go routine in the background waiting for owning process to be terminated, and when it happens
// it invokes defined Handlers sequentially. Every invocation will reset hooks.
func Listen() {
	Reset()

	hooks.Lock()
	signal.Notify(hooks.signals, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	hooks.open = true
	hooks.Unlock()

	go func() {
		defer func() {
			hooks.Lock()
			signal.Stop(hooks.signals)
			hooks.open = false
			hooks.Unlock()
		}()

		if _, ok := <-hooks.signals; !ok {
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
	if hooks.open {
		close(hooks.signals)
		hooks.open = false
	}
	hooks.Unlock()
}
