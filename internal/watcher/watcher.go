package watcher

import (
	"bytes"
	"errors"
	"sync"
	"syscall"
	"unsafe"
)

var (
	ErrWatcherClosed  = errors.New("watcher is closed")
	ErrAlreadyWatched = errors.New("path is already being watched")
	ErrNotWatched     = errors.New("path is not being watched")
)

type Watcher struct {
	fd    int
	paths map[string]uint32

	Events chan Event
	Errors chan error
	done   chan struct{}

	mu   sync.Mutex
	once sync.Once
}

type Event struct {
	Wd   int32
	Mask uint32
	Name string
}

func New() (*Watcher, error) {
	fd, err := syscall.InotifyInit1(syscall.IN_CLOEXEC)
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		fd:    fd,
		paths: make(map[string]uint32),

		Events: make(chan Event, 128),
		Errors: make(chan error, 16),
		done:   make(chan struct{}),
	}
	go w.run()

	return w, nil
}

func (w *Watcher) run() {
	defer close(w.Events)
	defer close(w.Errors)

	var buffer [syscall.SizeofInotifyEvent * 1024]byte

	for {
		n, err := syscall.Read(w.fd, buffer[:])

		if err != nil {
			if err == syscall.EINTR {
				continue
			}
			if err == syscall.EBADF && w.isDone() {
				return
			}
			if w.isDone() {
				return
			}
			w.emitError(err)
			return
		}

		if w.isDone() {
			return
		}

		offset := 0
		for offset < n {
			if n-offset < syscall.SizeofInotifyEvent {
				break
			}

			ev := (*syscall.InotifyEvent)(unsafe.Pointer(&buffer[offset]))

			eventSize := syscall.SizeofInotifyEvent + int(ev.Len)
			if eventSize <= 0 || offset+eventSize > n {
				break
			}

			nameBytes := buffer[offset+syscall.SizeofInotifyEvent : offset+eventSize]
			nameBytes = bytes.TrimRight(nameBytes, "\x00")
			name := string(nameBytes)

			e := Event{
				Wd:   int32(ev.Wd),
				Mask: ev.Mask,
				Name: name,
			}

			select {
			case w.Events <- e:
			default:
			}

			offset += eventSize
		}
	}
}

func (w *Watcher) AddWatch(path string, flags uint32) error {
	if w.isDone() {
		return ErrWatcherClosed
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	if _, ok := w.paths[path]; ok {
		return ErrAlreadyWatched
	}

	wd, err := syscall.InotifyAddWatch(w.fd, path, flags)
	if err != nil {
		return err
	}

	w.paths[path] = uint32(wd)

	return nil
}

func (w *Watcher) RemoveWatch(path string) error {
	if w.isDone() {
		return ErrWatcherClosed
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	wd, ok := w.paths[path]
	if !ok {
		return ErrNotWatched
	}

	_, err := syscall.InotifyRmWatch(w.fd, wd)
	if err != nil {
		return err
	}

	delete(w.paths, path)

	return nil
}

func (w *Watcher) Close() {
	w.once.Do(func() {
		close(w.done)

		w.mu.Lock()
		for path, wd := range w.paths {
			_, _ = syscall.InotifyRmWatch(w.fd, wd)
			delete(w.paths, path)
		}
		w.mu.Unlock()

		_ = syscall.Close(w.fd)
	})
}

func (w *Watcher) emitError(err error) {
	select {
	case w.Errors <- err:
	default:
	}
}

func (w *Watcher) isDone() bool {
	select {
	case <-w.done:
		return true
	default:
		return false
	}
}
