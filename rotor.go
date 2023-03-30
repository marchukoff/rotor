// Package rotor implements rolling *io.WriteCloser* for file loggers
package rotor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Rotor implements io.WriteCloser.
type Rotor struct {
	errors    chan error
	events    chan event
	written   chan int
	file      *os.File
	options   *options
	directory string
	filename  string
	filepath  string
}

// New returns a *rotor.Rotor and error if any.
func New(filename string, directory string, ops ...Option) (*Rotor, error) {
	opts := options{delimiter: "-", extension: "log", postfix: "2006-01-02", makeDir: true}

	for _, o := range ops {
		o.apply(&opts)
	}

	r := &Rotor{directory: directory, filename: filename, options: &opts}
	r.filepath = r.makeFilepath()
	if opts.makeDir {
		err := r.makeDir()
		if err != nil {
			return nil, err
		}
	}

	name := r.makeFilepath()
	f, err := r.makeFile(name)
	if err != nil {
		return nil, err
	}

	r.file = f
	r.run()

	return r, nil
}

// Write implements Write (io.Writer).
func (r *Rotor) Write(p []byte) (n int, err error) {
	r.events <- event{eventType: eventWrite, data: p}
	n = <-r.written
	err = <-r.errors
	return n, err
}

// Close implements Close (io.Closer, io.WriteCloser...).
func (r *Rotor) Close() error {
	r.events <- event{eventType: eventClose}
	err := <-r.errors
	return err
}

// Sync commits the current contents of the file to stable storage.
// Typically, this means flushing the file system's in-memory copy of recently written data to disk.
func (r *Rotor) Sync() error {
	r.events <- event{eventType: eventSync}
	err := <-r.errors
	return err
}

func (r *Rotor) run() {
	events := make(chan event, 10)
	r.events = events

	errors := make(chan error, 1)
	r.errors = errors

	written := make(chan int, 1)
	r.written = written

	go func() {
		for e := range events {
			switch e.eventType {
			case eventWrite:
				n, err := r.write(e.data)
				r.written <- n
				r.errors <- err
			case eventSync:
				var err error
				if r.file != nil {
					err = r.file.Sync()
				}
				r.errors <- err
			case eventClose:
				var err error
				if r.file != nil {
					err = r.file.Close()
					r.file = nil
				}
				r.errors <- err
			}
		}
	}()
}

func (r *Rotor) write(p []byte) (int, error) {
	if r.file == nil {
		return os.Stderr.Write(p)
	}

	name := r.makeFilepath()
	if !strings.EqualFold(r.filepath, name) {
		f, err := r.makeFile(name)
		if err != nil {
			return 0, err
		}
		_ = r.file.Close()
		r.filepath = name
		r.file = f
	}

	return r.file.Write(p)
}

func (r *Rotor) makeFile(name string) (*os.File, error) {
	const modeFile = 0o644
	return os.OpenFile(name, os.O_RDWR|os.O_APPEND|os.O_CREATE, modeFile)
}

func (r *Rotor) makeDir() error {
	const modeDir = 0o755
	_, err := os.Stat(r.directory)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(r.directory, modeDir)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

func (r *Rotor) makeFilepath() string {
	name := fmt.Sprint(r.filename, r.options.delimiter, time.Now().Format(r.options.postfix),
		".", r.options.extension)
	return filepath.Join(r.directory, name)
}
