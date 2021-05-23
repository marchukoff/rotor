package rotor

import (
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	prefixTimeFormat = "2006-01-02"
	extensionFile    = "log"
	defaultSuffix    = "app"
	defaultPath      = "./log"
	keepFiles        = 30
	modeFile         = 0644
	modeDir          = 0755
)

type RotationWriter interface {
	io.WriteCloser
}

type Rotor struct {
	mu sync.Mutex

	file  *os.File
	timer *time.Timer
	total int
	name  string
	path  string
}

func NewRotor(opts ...Option) *Rotor {
	r := &Rotor{
		file:  os.Stderr,
		name:  defaultSuffix,
		path:  defaultPath,
		total: keepFiles,
	}
	for _, opt := range opts {
		opt(r)
	}

	r.rotate()
	return r
}

func (r *Rotor) Write(p []byte) (n int, err error) {
	r.mu.Lock()
	n, err = r.file.Write(p)
	r.mu.Unlock()
	return n, err
}

func (r *Rotor) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.file == nil {
		return nil
	}

	err := r.file.Close()
	return err
}

func (r *Rotor) Sync() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.file == nil {
		return nil
	}

	err := r.file.Sync()
	return err
}

func (r *Rotor) Stop() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.timer == nil {
		return true
	}

	b := r.timer.Stop()
	return b
}

func (r *Rotor) rotate() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.file != nil {
		r.file.Sync()
		r.file.Close()
		r.file = nil
	}
	r.file = getFile(r.getFullFileName())

	if r.timer != nil {
		r.timer.Stop()
		r.timer = nil
	}
	r.timer = time.AfterFunc(getTimeToRotation(), r.rotate)

	go r.removeOldFiles()
}

func (r *Rotor) removeOldFiles() {
	all, err := ioutil.ReadDir(r.path)
	if err != nil {
		return
	}
	files := make([]fs.FileInfo, 0, len(all))
	suffix := strings.Join([]string{r.name, extensionFile}, ".")
	for _, file := range all {
		if strings.HasSuffix(file.Name(), suffix) {
			files = append(files, file)
		}
	}

	idx := len(files) - r.total
	if idx > 0 {
		for _, file := range files[0:idx] {
			os.Remove(filepath.Join(r.path, file.Name()))
		}
	}
}

func (r *Rotor) getFullFileName() string {
	prefix := time.Now().Format(prefixTimeFormat)
	name := strings.Join([]string{prefix, r.name, extensionFile}, ".")
	return filepath.Join(r.path, name)
}

func getTimeToRotation() time.Duration {
	now := time.Now()
	midnight := now.Truncate(24*time.Hour).AddDate(0, 0, 1).Add(time.Second)
	return midnight.Sub(now)
}

func getFile(filename string) *os.File {
	var (
		file *os.File
		err  error
	)
	dir := filepath.Dir(filename)
	if _, err = os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, modeDir); err != nil {
			fmt.Fprintf(os.Stderr, "error: cannot create log path for %q %v", filename, err)
			return os.Stderr
		}
	}

	if _, err = os.Stat(filename); os.IsNotExist(err) {
		file, err = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, modeFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: cannot create file %q %v", filename, err)
			return os.Stderr
		}
		return file
	}

	if file, err = os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, modeFile); err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot open file %q %v", filename, err)
		return os.Stderr
	}
	return file
}
