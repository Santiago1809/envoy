package watcher

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

const (
	DefaultDebounce = 50 * time.Millisecond
)

type Event struct {
	Path string
	Op   fsnotify.Op
	Time time.Time
}

type Watcher struct {
	watcher  *fsnotify.Watcher
	files    map[string]*fileState
	mu       sync.RWMutex
	onChange func(Event)
	onError  func(error)
	debounce time.Duration
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	closed   bool
}

type fileState struct {
	lastOp   fsnotify.Op
	lastTime time.Time
}

func New() (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	w := &Watcher{
		watcher:  fsw,
		files:    make(map[string]*fileState),
		debounce: DefaultDebounce,
		ctx:      ctx,
		cancel:   cancel,
	}

	w.wg.Add(1)
	go w.run()

	return w, nil
}

func (w *Watcher) SetDebounce(d time.Duration) {
	w.debounce = d
}

func (w *Watcher) OnChange(fn func(Event)) {
	w.onChange = fn
}

func (w *Watcher) OnError(fn func(error)) {
	w.onError = fn
}

func (w *Watcher) Add(path string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	stat, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat path: %w", err)
	}

	if stat.IsDir() {
		return w.watcher.Add(path)
	}

	w.files[path] = &fileState{}
	return w.watcher.Add(path)
}

func (w *Watcher) Remove(path string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	delete(w.files, path)
	return w.watcher.Remove(path)
}

func (w *Watcher) run() {
	defer w.wg.Done()

	for {
		select {
		case <-w.ctx.Done():
			return
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			w.handleEvent(event)
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			if w.onError != nil {
				w.onError(err)
			}
		}
	}
}

func (w *Watcher) handleEvent(event fsnotify.Event) {
	path := event.Name

	w.mu.RLock()
	state, exists := w.files[path]
	w.mu.RUnlock()

	now := time.Now()

	if exists {
		if now.Sub(state.lastTime) < w.debounce {
			state.lastOp = event.Op
			state.lastTime = now
			return
		}
		state.lastOp = event.Op
		state.lastTime = now
	}

	if event.Op&fsnotify.Create == fsnotify.Create ||
		event.Op&fsnotify.Write == fsnotify.Write ||
		event.Op&fsnotify.Remove == fsnotify.Remove ||
		event.Op&fsnotify.Rename == fsnotify.Rename {

		e := Event{
			Path: path,
			Op:   event.Op,
			Time: now,
		}

		if w.onChange != nil {
			w.onChange(e)
		}
	}
}

func (w *Watcher) Start(path string) error {
	return w.Add(path)
}

func (w *Watcher) Stop() error {
	w.cancel()
	w.wg.Wait()

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return nil
	}

	w.closed = true
	return w.watcher.Close()
}

func (w *Watcher) Watch(path string, onChange func(Event)) error {
	w.OnChange(onChange)
	return w.Add(path)
}

func (w *Watcher) WatchWithExec(path string, execCmd string, onChange func(Event)) error {
	w.OnChange(onChange)
	return w.Add(path)
}

func WatchFile(path string, onChange func(Event), debounce time.Duration) (*Watcher, error) {
	w, err := New()
	if err != nil {
		return nil, err
	}

	if debounce > 0 {
		w.SetDebounce(debounce)
	}

	w.OnChange(onChange)

	err = w.Add(path)
	if err != nil {
		w.Stop()
		return nil, err
	}

	return w, nil
}

func WatchWithSignal(path string, onChange func(Event), debounce time.Duration) error {
	w, err := WatchFile(path, onChange, debounce)
	if err != nil {
		return err
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill)

	select {
	case <-sigChan:
		fmt.Println("\nShutting down...")
		w.Stop()
	}

	return nil
}

func WaitForFile(path string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for file: %w", ctx.Err())
		case <-ticker.C:
			if _, err := os.Stat(path); err == nil {
				return nil
			}
		}
	}
}
