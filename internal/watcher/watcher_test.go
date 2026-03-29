package watcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	w, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer w.Stop()

	if w.watcher == nil {
		t.Error("watcher should not be nil")
	}
}

func TestAdd(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	w, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer w.Stop()

	err = w.Add(tmpfile.Name())
	if err != nil {
		t.Errorf("Add() error = %v", err)
	}
}

func TestRemove(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	w, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer w.Stop()

	err = w.Add(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	err = w.Remove(tmpfile.Name())
	if err != nil {
		t.Errorf("Remove() error = %v", err)
	}
}

func TestDebounce(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	w, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer w.Stop()

	w.SetDebounce(100 * time.Millisecond)

	eventCount := 0
	w.OnChange(func(e Event) {
		eventCount++
	})

	err = w.Add(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 5; i++ {
		tmpfile, err = os.OpenFile(tmpfile.Name(), os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			t.Fatal(err)
		}
		tmpfile.WriteString("test")
		tmpfile.Close()
		time.Sleep(10 * time.Millisecond)
	}

	time.Sleep(200 * time.Millisecond)

	if eventCount == 0 {
		t.Error("expected at least one event")
	}
}

func TestOnChange(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	w, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer w.Stop()

	events := make([]Event, 0)
	w.OnChange(func(e Event) {
		events = append(events, e)
	})

	err = w.Add(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	tmpfile, err = os.OpenFile(tmpfile.Name(), os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatal(err)
	}
	tmpfile.WriteString("content")
	tmpfile.Close()

	time.Sleep(100 * time.Millisecond)

	if len(events) == 0 {
		t.Error("expected at least one event")
	}
}

func TestWatchFile(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	eventCalled := false

	w, err := WatchFile(tmpfile.Name(), func(e Event) {
		eventCalled = true
	}, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("WatchFile() error = %v", err)
	}
	defer w.Stop()

	tmpfile, err = os.OpenFile(tmpfile.Name(), os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatal(err)
	}
	tmpfile.WriteString("test")
	tmpfile.Close()

	time.Sleep(100 * time.Millisecond)

	if !eventCalled {
		t.Error("expected event to be called")
	}
}

func TestWaitForFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "newfile.txt")

	go func() {
		time.Sleep(50 * time.Millisecond)
		os.WriteFile(path, []byte("test"), 0644)
	}()

	err := WaitForFile(path, 2*time.Second)
	if err != nil {
		t.Errorf("WaitForFile() error = %v", err)
	}
}

func TestWaitForFileTimeout(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.txt")

	err := WaitForFile(path, 100*time.Millisecond)
	if err == nil {
		t.Error("expected timeout error")
	}
}

func TestOnError(t *testing.T) {
	w, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer w.Stop()

	errorCalled := false
	w.OnError(func(e error) {
		errorCalled = true
	})

	_ = errorCalled
}

func TestStop(t *testing.T) {
	w, err := New()
	if err != nil {
		t.Fatal(err)
	}

	err = w.Stop()
	if err != nil {
		t.Errorf("Stop() error = %v", err)
	}

	err = w.Stop()
	if err != nil {
		t.Errorf("Stop() second call error = %v", err)
	}
}

func TestWatchDirectory(t *testing.T) {
	dir := t.TempDir()

	w, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer w.Stop()

	events := make([]Event, 0)
	w.OnChange(func(e Event) {
		events = append(events, e)
	})

	err = w.Add(dir)
	if err != nil {
		t.Fatal(err)
	}

	tmpfile := filepath.Join(dir, "test.txt")
	os.WriteFile(tmpfile, []byte("test"), 0644)
	defer os.Remove(tmpfile)

	time.Sleep(100 * time.Millisecond)

	if len(events) == 0 {
		t.Error("expected event for file creation in directory")
	}
}
