package locker

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"

	"github.com/spf13/afero"
)

type Locker struct {
	fs      afero.Fs
	path    string
	content interface{}
}

func New(path string, content interface{}) *Locker {
	return _new(afero.OsFs{}, path, content)
}

func _new(from afero.Fs, path string, content interface{}) *Locker {
	return &Locker{
		fs:      from,
		path:    path,
		content: content,
	}
}

func (l Locker) ShutdownContext(ctx context.Context) {
	go func() {
		select {
		case <-ctx.Done():
			_ = l.Unlock()
		}
	}()
}

// ErrorAlreadyLocked reports that our path is already locked.
var ErrorAlreadyLocked = errors.New("LockPath is already locked")

func (l Locker) exists() bool {
	_, err := l.fs.Stat(l.path)
	return !os.IsNotExist(err)
}

// lock creates a filesystem lock for a process
func (l Locker) Lock() error {
	if l.exists() {
		return ErrorAlreadyLocked
	}

	file, err := l.fs.OpenFile(l.path, os.O_RDWR|os.O_CREATE, 0744)
	if err != nil {
		return err
	}
	text, err := json.Marshal(l.content)
	if err != nil {
		return err
	}
	_, err = file.Write(text)
	return err
}

// unlock removes a lock file if it exists
func (l Locker) Unlock() error {
	err := l.fs.Remove(l.path)
	if os.IsNotExist(err) {
		err = nil
	}
	return err
}

// Read reads a lockfile to get its content
func (l Locker) Read(val interface{}) error {
	file, err := l.fs.Open(l.path)
	// we have an error but it's not what we're expecting.
	if err != nil {
		return err
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	// Convert the file contents to its interface target.
	return json.Unmarshal(data, val)
}
