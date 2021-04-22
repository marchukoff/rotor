package rotor

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	defaultFilePrefix    = "2006-01-02"
	defaultFileExtension = "log"
	defaultFilePath      = "log"
	defaultDirMode       = 0755
	defaultFileMode      = 0644
	defaultFileAge       = time.Hour * 24 * 7
)

var _ io.WriteCloser = (*Rotor)(nil)

type Rotor struct {
	mu sync.Mutex

	timer         *time.Timer
	file          *os.File
	fileName      string
	filePath      string
	fileExtension string
	filePrefix    string
	fileAge       time.Duration
}

func NewRotor(opts ...Option) (*Rotor, error) {
	rot := &Rotor{
		fileName:      "app",
		filePath:      defaultFilePath,
		fileExtension: defaultFileExtension,
		fileAge:       defaultFileAge,
	}
	for _, opt := range opts {
		if err := opt(rot); err != nil {
			return rot, err
		}
	}
	rot.timer = time.AfterFunc(getTimeToNewRotation(), func() { rot.rotate() })
	return rot, nil
}

type Option func(rot *Rotor) error

func WithFilePath(path string) Option {
	return func(rot *Rotor) error {
		if p, err := filepath.Abs(filepath.Clean(path)); err != nil {
			return err
		} else {
			rot.filePath = p
		}
		if _, err := os.Stat(rot.filePath); os.IsNotExist(err) {
			err = os.MkdirAll(rot.filePath, defaultDirMode)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func WithFileName(name string) Option {
	return func(rot *Rotor) error {
		rot.fileName = name
		return nil
	}
}

func WithFileExtension(extension string) Option {
	return func(rot *Rotor) error {
		rot.fileExtension = extension
		return nil
	}
}

func WithFileAge(age time.Duration) Option {
	return func(rot *Rotor) error {
		rot.fileAge = age
		return nil
	}
}

func (r *Rotor) Write(p []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.file == nil {
		if err = r.openExistingOrNew(); err != nil {
			return 0, err
		}
	}
	return r.file.Write(p)
}

func (r *Rotor) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.close()
}

func (r *Rotor) close() error {
	if r.file == nil {
		return nil
	}
	err := r.file.Close()
	r.file = nil
	return err
}

func (r *Rotor) Sync() error {
	if r.file == nil {
		return nil
	}
	return r.file.Sync()
}

func (r *Rotor) rotate() {
	go func() {
		files, err := ioutil.ReadDir(r.filePath)
		if err == nil {
			maxAge := time.Now().Add(-r.fileAge)
			for _, file := range files {
				if file.ModTime().Before(maxAge) {
					_ = os.Remove(filepath.Join(r.filePath, file.Name()))
				}
			}
		}
	}()

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.timer != nil {
		r.timer.Stop()
		r.timer = nil
	}

	_ = r.Sync()
	_ = r.close()
	r.file = nil
	r.timer = time.AfterFunc(getTimeToNewRotation(), func() { r.rotate() })
}

func (r *Rotor) filename() string {
	return filepath.Join(
		r.filePath,
		strings.Join(
			[]string{time.Now().Format(defaultFilePrefix), r.fileName, r.fileExtension},
			".",
		),
	)
}

func (r *Rotor) openExistingOrNew() error {
	filename := r.filename()
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return r.openNew()
	}
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, defaultFileMode)
	if err != nil {
		return r.openNew()
	}
	r.file = file
	return nil
}

func (r *Rotor) openNew() error {
	if err := os.MkdirAll(r.filePath, defaultDirMode); err != nil {
		return err
	}
	filename := r.filename()
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, defaultFileMode)
	if err != nil {
		return err
	}
	r.file = file
	return nil
}

func getTimeToNewRotation() time.Duration {
	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	return time.Second + midnight.Sub(now)
}
