//go:build windows

package events

type windowsSources struct {
	hook *winEventHook
	done chan struct{}
}

func newWindowsSources(emit func(SystemEvent)) *windowsSources {
	return &windowsSources{
		hook: &winEventHook{emit: emit, ready: make(chan struct{})},
		done: make(chan struct{}),
	}
}

func (w *windowsSources) start() {
	go func() {
		defer close(w.done)
		w.hook.run()
	}()
	<-w.hook.ready
}

func (w *windowsSources) stop() {
	w.hook.stop()
	<-w.done
}
