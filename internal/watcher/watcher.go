package watcher

import (
	"errors"
	"sync"
	"syscall"
)

type Watcher struct {
	fd    int
	paths map[string]uint32
	done  chan struct{}

	mu   sync.Mutex
	once sync.Once
	wg   sync.WaitGroup
}

func New() (*Watcher, error) {
	fd, err := syscall.InotifyInit1(syscall.IN_CLOEXEC)
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		fd:    fd,
		paths: make(map[string]uint32),
		done:  make(chan struct{}),
	}
	w.wg.Add(1)
	go w.run()

	return w, nil
}

func (w *Watcher) run() {
	defer w.wg.Done()

	var buffer [syscall.SizeofInotifyEvent * 1024]byte

	for {
		n, err := syscall.Read(w.fd, buffer[:])
		_ = n

		if err != nil {
			if err == syscall.EINTR {
				continue
			}

			if isDone(w.done) {
				return
			}
			return
		}

		if isDone(w.done) {
			return
		}
	}
}

func (w *Watcher) AddWatch(path string, flags uint32) error {
	if isDone(w.done) {
		return errors.New("watcher is closed")
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	if _, ok := w.paths[path]; ok {
		return errors.New("file is already being watched")
	}

	wd, err := syscall.InotifyAddWatch(w.fd, path, flags)
	if err != nil {
		return err
	}

	w.paths[path] = uint32(wd)

	return nil
}

func (w *Watcher) RemoveWatch(path string) error {
	if isDone(w.done) {
		return errors.New("watcher is closed")
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	wd, ok := w.paths[path]
	if !ok {
		return errors.New("file is not being watched")
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

	w.wg.Wait()
}

func isDone(done <-chan struct{}) bool {
	select {
	case <-done:
		return true
	default:
		return false
	}
}
