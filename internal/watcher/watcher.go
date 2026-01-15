package watcher

import (
	"errors"
	"syscall"
)

type Watcher struct {
	fd    int
	paths map[string]uint32
}

func New() (*Watcher, error) {
	fd, err := syscall.InotifyInit1(syscall.IN_CLOEXEC)
	if err != nil {
		return nil, err
	}

	return &Watcher{
		fd:    fd,
		paths: make(map[string]uint32),
	}, nil
}

func (w *Watcher) AddWatch(path string, flags uint32) error {
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
	for path := range w.paths {
		w.RemoveWatch(path)
	}
	syscall.Close(w.fd)
}

// syscall.Read
